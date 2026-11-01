import {test} from 'node:test';
import assert from 'node:assert/strict';
import {JSDOM} from 'jsdom';
import {Reader} from '../build/features/reading/index.js';
const tick=()=>new Promise(resolve=>setImmediate(resolve));
const config={model:'synthetic-model',target:'zh-CN',locale:'zh',automatic:false};
function environment() {
  const dom=new JSDOM('<main><p>Hello world.</p></main>');
  globalThis.Element=dom.window.Element;globalThis.MutationObserver=dom.window.MutationObserver;
  return dom;
}
test('manual injection does not also start a saved automatic rule',async()=>{
  const dom=environment();let calls=0;
  const reader=new Reader(dom.window.document,async type=>{
    if(type==='contentConfig')return {...config,automatic:true};
    if(type==='translateBlock'){calls++;return 'Synthetic translation';}
  });
  await reader.initialize(false);
  assert.equal(calls,0);
  dom.window.close();
});
test('reader inserts bilingual text once and restores the same original paragraph',async()=>{
  const dom=environment(),document=dom.window.document,p=document.querySelector('p');
  let calls=0;
  const reader=new Reader(document,async type=>{
    if(type==='contentConfig')return config;
    if(type==='translateBlock'){calls++;return '你好，世界。';}
  });
  await reader.initialize();assert.equal(calls,0);
  await reader.action('toggle');
  assert.equal(calls,1);assert.ok(p.querySelector('[data-shadowinput]'));
  await reader.action('toggle');
  assert.equal(p.querySelector('[data-shadowinput]'),null);
  assert.equal(document.querySelector('p'),p);
  dom.window.close();
});
test('cancel then immediate retry owns a new run; the old response cannot stop or annotate it',async()=>{
  const dom=environment(),document=dom.window.document,replies=[];
  const reader=new Reader(document,async type=>{
    if(type==='contentConfig')return config;
    if(type==='translateBlock')return new Promise(resolve=>replies.push(resolve));
  });
  const first=reader.action('toggle');await tick();
  await reader.action('toggle');
  const second=reader.action('toggle');await tick();
  assert.equal(replies.length,2);
  replies[0]('Old translation');await first;
  assert.equal(document.querySelector('[data-shadowinput="translation"]'),null);
  replies[1]('New translation');await second;
  assert.ok(document.querySelector('[data-shadowinput="translation"]').shadowRoot.textContent.includes('New translation'));
  await reader.action('toggle');dom.window.close();
});
