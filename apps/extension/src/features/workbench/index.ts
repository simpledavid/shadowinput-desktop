import {copy,type Locale} from '../../shared/copy.js';
import {errorCode} from '../../shared/errors.js';
import {request} from '../../shared/rpc.js';
import {object,requiredString} from '../../shared/validation.js';
import {settings} from '../../shared/settings.js';
import {diagnosticOf,diagnosticText} from '../../shared/diagnostics.js';

function element<T extends HTMLElement>(id: string, constructor: {new(): T}): T {
  const found = document.getElementById(id);
  if (!(found instanceof constructor)) throw new Error('missing_element');
  return found;
}

export function mountWorkbench(): void {
  const locale: Locale = navigator.language.startsWith('zh') ? 'zh' : 'en';
  const language = element('language',HTMLSelectElement);
  language.value=locale;
  let t = copy[language.value === 'zh' ? 'zh' : 'en'];
  const source = element('source',HTMLTextAreaElement);
  const result = element('result',HTMLTextAreaElement);
  const model = element('model',HTMLSelectElement);
  const target = element('target',HTMLSelectElement);
  const generate = element('translate',HTMLButtonElement);
  const cancel = element('cancel',HTMLButtonElement);
  const copyButton = element('copy',HTMLButtonElement);
  const connect = element('connect',HTMLButtonElement);
  const logout = element('logout',HTMLButtonElement);
  const cancelSignIn = element('cancel-sign-in',HTMLButtonElement);
  const status = element('status',HTMLParagraphElement);
  const error = element('error',HTMLParagraphElement);
  let connected = false, busy = false, loadingModels = false, lastNotice = '';
  let savedModel='';

  async function save():Promise<void> {
    await request('saveSettings',{settings:{target:target.value,model:model.value,locale:language.value}});
  }

  function applyCopy(): void {
    t = copy[language.value === 'zh' ? 'zh' : 'en'];
    document.documentElement.lang = language.value === 'zh' ? 'zh-CN' : 'en';
    document.querySelectorAll<HTMLElement>('[data-copy]').forEach(node=>{
      const key = node.dataset.copy;
      for (const [name,value] of Object.entries(t)) if (name === key && typeof value === 'string') node.textContent = value;
    });
    language.ariaLabel = t.language;
    source.placeholder = t.placeholder; result.placeholder = t.empty;
    Array.from(target.options).forEach((option,index)=>{option.textContent=t.languages[index] ?? '';});
    lastNotice = '';
    updateButtons();
  }
  function updateButtons(): void {
    generate.disabled = !connected || busy || !model.value || !source.value.trim();
    generate.textContent = busy ? t.translating : t.translate;
    copyButton.disabled = !result.value || busy;
    cancel.hidden = !busy;
  }
  async function run(action:()=>Promise<void>): Promise<void> {
    error.textContent = '';
    try {await action();} catch (reason) {showError(reason);}
  }
  function showError(reason:unknown):void {
    error.textContent=t.errors[errorCode(reason)]+diagnosticText(diagnosticOf(reason),t.stages);
  }
  async function loadModels(): Promise<void> {
    if (loadingModels) return;
    loadingModels = true;
    try {
      const list = await request('models');
      if (!Array.isArray(list)) throw new Error('invalid_response');
      const selected = model.value || savedModel;
      const options = list.map(value=>{
        const item=object(value);
        return new Option(requiredString(item.name),requiredString(item.id));
      });
      model.replaceChildren(...options);
      if (options.some(option=>option.value === selected)) model.value=selected;
      await save();
      updateButtons();
    } finally { loadingModels=false; }
  }
  async function refresh(): Promise<void> {
    const data=object(await request('status'));
    const newlyConnected=!connected && data.connected === true;
    connected=data.connected === true;
    connect.hidden=connected || data.pending === true || data.notice === 'verifying';
    logout.hidden=!connected;
    cancelSignIn.hidden=data.pending !== true;
    status.textContent=data.notice === 'waiting' ? t.waiting : data.notice === 'verifying' ? t.verifying :
      connected ? (typeof data.email === 'string' && data.email ? data.email : t.connected) : t.signedOut;
    if (typeof data.notice === 'string' && lastNotice !== data.notice) {
      lastNotice=data.notice;
      if (data.notice in t.errors) error.textContent=t.errors[errorCode(new Error(data.notice))]+diagnosticText(data.diagnostic,t.stages);
    }
    updateButtons();
    if (newlyConnected) await loadModels();
  }
  language.addEventListener('change',()=>{applyCopy();void run(async()=>{await save();await refresh();});});
  connect.addEventListener('click',()=>void run(async()=>{await request('connect');await refresh();}));
  cancelSignIn.addEventListener('click',()=>void run(async()=>{await request('cancelSignIn');await refresh();}));
  logout.addEventListener('click',()=>void run(async()=>{await request('logout');model.replaceChildren();result.value='';await refresh();}));
  element('refresh',HTMLButtonElement).addEventListener('click',()=>void run(loadModels));
  source.addEventListener('input',updateButtons);
  model.addEventListener('change',()=>{updateButtons();void run(save);});
  target.addEventListener('change',()=>void run(save));
  element('translation-form',HTMLFormElement).addEventListener('submit',event=>{
    event.preventDefault();
    if (generate.disabled) return;
    void run(async()=>{
      busy=true;result.value='';copyButton.textContent=t.copy;updateButtons();
      try {
        result.value=requiredString(await request('translate',{text:source.value,target:target.value,model:model.value}));
      } finally {busy=false;updateButtons();}
    });
  });
  cancel.addEventListener('click',()=>void run(async()=>{await request('cancel');}));
  copyButton.addEventListener('click',()=>void run(async()=>{await navigator.clipboard.writeText(result.value);copyButton.textContent=t.copied;}));
  applyCopy();
  void run(async()=>{
    const config=settings(await request('settings'));
    language.value=config.locale;target.value=config.target;savedModel=config.model;applyCopy();
    await refresh();
  });
  const timer=setInterval(()=>void refresh().catch(showError),2000);
  window.addEventListener('pagehide',()=>clearInterval(timer),{once:true});
}
