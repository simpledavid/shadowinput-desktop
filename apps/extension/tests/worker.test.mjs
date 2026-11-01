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
