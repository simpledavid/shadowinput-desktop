import {copy} from '../../shared/copy.js';
import {readingCopy} from '../../shared/reading-copy.js';
import {request} from '../../shared/rpc.js';
import {errorCode} from '../../shared/errors.js';
import {settings} from '../../shared/settings.js';
import {object,requiredString} from '../../shared/validation.js';
import {diagnosticOf,diagnosticText} from '../../shared/diagnostics.js';

export async function mountToolbar():Promise<void> {
  const config=settings(await request('settings'));
  const text=readingCopy[config.locale],t=copy[config.locale];
  document.documentElement.lang=config.locale==='zh'?'zh-CN':'en';
  const status=document.querySelector('p[role="status"]'),actions=document.querySelector('section');
  const target=document.querySelector('#target'),model=document.querySelector('#model');
  const automatic=document.querySelector('#automatic');
  if(!(status instanceof HTMLParagraphElement) || !(actions instanceof HTMLElement) ||
    !(target instanceof HTMLSelectElement) || !(model instanceof HTMLSelectElement) || !(automatic instanceof HTMLInputElement))throw new Error('invalid_response');
  const targetField=target,modelField=model;
  document.querySelectorAll<HTMLElement>('[data-copy]').forEach(node=>{
    for(const [key,value] of Object.entries({...text,...t}))if(key===node.dataset.copy && typeof value==='string')node.textContent=value;
  });
  Array.from(target.options).forEach((option,index)=>{option.textContent=t.languages[index] ?? '';});
  target.value=config.target;
  async function run(action:()=>Promise<void>) {
    status!.textContent='';
    try{await action();}catch(error){status!.textContent=t.errors[errorCode(error)]+diagnosticText(diagnosticOf(error),t.stages);}
  }
  const account=object(await request('status'));
  const login=document.querySelector('#connect');
  if(!(login instanceof HTMLButtonElement))throw new Error('invalid_response');
  login.hidden=account.connected===true;
  actions.hidden=account.connected!==true;
  login.addEventListener('click',()=>void run(async()=>{await request('connect');window.close();}));
  document.querySelector('#settings')?.addEventListener('click',()=>{void chrome.tabs.create({url:chrome.runtime.getURL('panel.html')});window.close();});
  let host:string|undefined;
  await run(async()=>{
    const site=object(await request('currentSite'));host=requiredString(site.origin);automatic.checked=site.automatic===true;
    const hostLabel=document.querySelector('#host');if(hostLabel)hostLabel.textContent=new URL(host).hostname;
  });
  automatic.disabled=!host;
  automatic.addEventListener('change',()=>{
    const enabled=automatic.checked;
    if(!host)return;
    const chosenHost=host;
    // Permission is requested directly from this user gesture, for one origin.
    const granted=enabled?chrome.permissions.request({origins:[chosenHost+'/*']}):Promise.resolve(true);
    void run(async()=>{
      if(!await granted){automatic.checked=false;throw new Error('site_permission');}
      await request('setAutomatic',{origin:chosenHost,enabled});
      status!.textContent=text.save;
    });
  });
  async function save() {
    await request('saveSettings',{settings:{...config,target:targetField.value,model:modelField.value}});
  }
  target.addEventListener('change',()=>void run(save));
  model.addEventListener('change',()=>void run(save));
  document.querySelectorAll<HTMLButtonElement>('[data-action]').forEach(button=>{
    button.addEventListener('click',()=>void run(async()=>{await save();await request('pageAction',{action:button.dataset.action});window.close();}));
  });
  if(account.connected===true)await run(async()=>{
    const list=await request('models');
    if(!Array.isArray(list))throw new Error('invalid_response');
    const options=list.map(value=>{const item=object(value);return new Option(requiredString(item.name),requiredString(item.id));});
    model.replaceChildren(...options);
    if(options.some(option=>option.value===config.model))model.value=config.model;
    await save();
  });
}
