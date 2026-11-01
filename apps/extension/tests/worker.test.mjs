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
