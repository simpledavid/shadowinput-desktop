import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {JSDOM} from 'jsdom';
import {mountToolbar} from '../build/features/toolbar/index.js';

const html=await readFile(new URL('../public/popup.html',import.meta.url),'utf8');
async function toolbar(supported) {
  const dom=new JSDOM(html,{url:'https://fixture.invalid/'}),calls=[];
  for(const name of ['document','window','HTMLElement','HTMLParagraphElement','HTMLSelectElement','HTMLInputElement','HTMLButtonElement','Option'])globalThis[name]=dom.window[name];
  globalThis.chrome={runtime:{async sendMessage(message){
    calls.push(message.type);
    if(message.type==='currentSite' && !supported)return {ok:false,error:'unsupported_page'};
    const values={settings:{target:'zh-CN',locale:'zh',model:'test-model'},status:{connected:true},
      currentSite:{origin:'https://example.invalid',automatic:false},models:[{id:'test-model',name:'Test model'}]};
    return {ok:true,data:values[message.type]};
  }}};
  await mountToolbar();
  return {dom,calls};
}

test('the account page disables webpage actions and shows a neutral next step',async()=>{
  const {dom,calls}=await toolbar(false);
  try {
    const buttons=[...dom.window.document.querySelectorAll('[data-action]')];
    assert.ok(buttons.every(button=>button.disabled),'unsupported-page actions must not be clickable');
    for(const button of buttons)button.click();
    assert.ok(!calls.includes('pageAction'));
    assert.equal(dom.window.document.querySelector('.site-setting').hidden,true);
    const status=dom.window.document.querySelector('[role="status"]');
    assert.equal(status.textContent,'请切换到需要翻译的网页，再打开插件。');
    assert.equal(status.classList.contains('error'),false);
  } finally {dom.window.close();}
});

test('a supported webpage enables translation after its model is loaded',async()=>{
  const {dom}=await toolbar(true);
  try {
    assert.ok([...dom.window.document.querySelectorAll('[data-action]')].every(button=>!button.disabled));
    assert.equal(dom.window.document.querySelector('.site-setting').hidden,false);
    assert.equal(dom.window.document.querySelector('#automatic').disabled,false);
  } finally {dom.window.close();}
});
