import {test} from 'node:test';
import assert from 'node:assert/strict';
import {begin} from '../build/features/connect/index.js';

const events={};
const event=name=>({addListener(fn){events[name]=fn;}});
const local={},session={};
const area=data=>({
  async setAccessLevel(){},
  async get(key){if(typeof key==='string')return {[key]:data[key]};return {...data};},
  async set(next){Object.assign(data,next);},
  async remove(key){delete data[key];},
});
const navigations=[];
globalThis.fetch=async()=>{throw new TypeError('private-transport-canary');};
globalThis.chrome={
  storage:{local:area(local),session:area(session)},
  runtime:{id:'test-extension',getURL:path=>'chrome-extension://test-extension/'+path,onMessage:event('message'),onInstalled:event('installed')},
  tabs:{onRemoved:event('removed'),async update(id,options){navigations.push({id,options});}},
  webNavigation:{onBeforeNavigate:event('navigate'),onCommitted:event('committed')},
  commands:{onCommand:event('command')},contextMenus:{onClicked:event('context')},
};
await import('../build/app/background.js');
const external={id:'another-extension',url:'https://example.invalid/',frameId:0,documentId:'external',tab:{id:1}};
const content={...external,id:'test-extension',documentId:'content'};
const tick=()=>new Promise(resolve=>setImmediate(resolve));

test('a page or another extension cannot read credentials or invoke account operations',()=>{
  for(const sender of [external,content]) {
    for(const type of ['status','connect','logout','models','settings','translate','setAutomatic']) {
      let called=false;
      assert.equal(events.message({type},sender,()=>{called=true;}),false);
      assert.equal(called,false);
    }
  }
});
test('only the owned top-level authorization tab can consume the pending callback',async()=>{
  const attempt=await begin({hostID:'urn:uuid:00000000-0000-4000-8000-000000000001'});
  session.pending={tabID:42,attempt};
  const url=attempt.redirect+'?'+new URLSearchParams({state:attempt.state,code:'synthetic-code',client_id:'oaiapp_test'});
  events.navigate({tabId:7,frameId:0,url});await tick();
  events.navigate({tabId:42,frameId:2,url});await tick();
  assert.ok(session.pending);
  assert.equal(navigations.length,0);
  events.navigate({tabId:42,frameId:0,url:attempt.redirect+'?state=forged&code=x&client_id=oaiapp_test'});await tick();await tick();
  assert.equal(session.pending,undefined);
  assert.equal(session.notice,'invalid_callback');
  assert.deepEqual(navigations,[{id:42,options:{url:'chrome-extension://test-extension/panel.html'}}]);
  events.navigate({tabId:42,frameId:0,url});await tick();
  assert.equal(navigations.length,1,'a consumed callback cannot be replayed');
});

test('a failed callback exposes only safe stage diagnostics to the workbench',async()=>{
  const record={hostID:'urn:uuid:00000000-0000-4000-8000-000000000002'};
  local.account=record;
  const attempt=await begin(record);
  session.pending={tabID:43,attempt};
  const url=attempt.redirect+'?'+new URLSearchParams({state:attempt.state,code:'private-code-canary',client_id:'oaiapp_test'});
  events.navigate({tabId:43,frameId:0,url});
  await tick();await tick();await tick();
  const sender={id:'test-extension',url:'chrome-extension://test-extension/panel.html'};
  const reply=await new Promise(resolve=>events.message({type:'status'},sender,resolve));
  assert.equal(reply.ok,true);
  assert.equal(reply.data.connected,false);
  assert.deepEqual(reply.data.diagnostic,{stage:'token_exchange',code:'network_error',kind:'transport'});
  assert.doesNotMatch(JSON.stringify(reply),/private-|verifier|access_token/);
});

test('an old model request cannot sign out a newly stored account session',async()=>{
  local.account={hostID:'host',clientID:'oaiapp_test',subject:'synthetic-user',session:{
    email:'test@example.invalid',access:'old-access',refresh:'old-refresh',scope:'chatgpt.tokens.use.direct',expiresAt:Date.now()+3600000,
  }};
  const original=globalThis.fetch;
  let rejectRequest,started;
  const sent=new Promise(resolve=>{started=resolve;});
  globalThis.fetch=async()=>{started();return new Promise(resolve=>{rejectRequest=resolve;});};
  try {
    const sender={id:'test-extension',url:'chrome-extension://test-extension/panel.html'};
    const response=new Promise(resolve=>events.message({type:'models'},sender,resolve));
    await sent;
    local.account.session={...local.account.session,access:'new-access',refresh:'new-refresh'};
    rejectRequest(new Response('',{status:401}));
    assert.equal((await response).error,'sign_in_required');
    const status=await new Promise(resolve=>events.message({type:'status'},sender,resolve));
    assert.equal(status.data.connected,true,'late old-session errors must preserve the new login');
    assert.equal(local.account.session.access,'new-access');
  } finally {globalThis.fetch=original;}
});

test('a current session rejected by the provider becomes signed out',async()=>{
  const original=globalThis.fetch;
  globalThis.fetch=async()=>new Response('',{status:401});
  try {
    local.account.session={email:'test@example.invalid',access:'current',refresh:'refresh',scope:'chatgpt.tokens.use.direct',expiresAt:Date.now()+3600000};
    const sender={id:'test-extension',url:'chrome-extension://test-extension/panel.html'};
    const reply=await new Promise(resolve=>events.message({type:'models'},sender,resolve));
    assert.equal(reply.error,'sign_in_required');
    assert.equal(local.account.session,undefined);
  } finally {globalThis.fetch=original;}
});

test('closing the text workbench cancels its in-flight translation',async()=>{
  const original=globalThis.fetch;
  let signal,release,started;
  const sent=new Promise(resolve=>{started=resolve;});
  globalThis.fetch=async(_url,options)=>{signal=options.signal;started();return new Promise(resolve=>{release=resolve;});};
  local.account.session={email:'test@example.invalid',access:'current',refresh:'refresh',scope:'chatgpt.tokens.use.direct',expiresAt:Date.now()+3600000};
  const sender={id:'test-extension',url:'chrome-extension://test-extension/panel.html',documentId:'workbench',frameId:0,tab:{id:88}};
  const response=new Promise(resolve=>events.message({type:'translate',text:'Synthetic sentence.',target:'zh-CN',model:'test-model'},sender,resolve));
  await sent;
  try {
    events.removed(88);await tick();
    assert.equal(signal.aborted,true,'closing the owner should cancel provider work');
  } finally {
    release(new Response(''));
    await response;
    globalThis.fetch=original;
  }
});
