import {Account} from '../features/connect/index.js';
import {Jobs,models,translate} from '../features/translate/index.js';
import {provider} from '../shared/config.js';
import {errorCode} from '../shared/errors.js';
import {diagnosticOf,parseDiagnostic,type Diagnostic} from '../shared/diagnostics.js';
import {object,requiredString} from '../shared/validation.js';
import {store,ready} from './store.js';
import {pending} from './pending.js';
import {preferences,savePreferences,currentSite,setAutomatic,contentConfig,sendAction,origin} from './sites.js';

const account = new Account(store);
const page = chrome.runtime.getURL('panel.html');
let control: Promise<unknown> = Promise.resolve();
const jobs=new Jobs();
const documentTabs=new Map<string,number>();

function ordered<T>(action:()=>Promise<T>):Promise<T> {
  const result = control.then(action);
  control = result.then(()=>undefined,()=>undefined);
  return result;
}

async function notice(value: string,diagnostic?:Diagnostic): Promise<void> {
  await chrome.storage.session.set({notice:value,diagnostic:diagnostic ?? null});
}

async function start(): Promise<void> {
  const saved = await chrome.storage.session.get('pending');
  if (pending(saved.pending)) throw new Error('busy');
  const attempt = await account.start();
  const tab = await chrome.tabs.create({url:'about:blank',active:true});
  if (tab.id === undefined) throw new Error('network_error');
  await chrome.storage.session.set({pending:{attempt,tabID:tab.id},notice:'waiting',diagnostic:null});
  await chrome.tabs.update(tab.id,{url:attempt.authorizationURL});
}

async function finish(tabID: number, url: string): Promise<void> {
  await ready;
  const saved = await chrome.storage.session.get('pending');
  const active = pending(saved.pending);
  // Only this extension's newly created top-level authorization tab owns a callback.
  if (!active || active.tabID !== tabID) return;
  const attempt = active.attempt;
  if (!url.startsWith(attempt.redirect+'?')) return;
  await chrome.storage.session.remove('pending');
  await notice('verifying');
  await chrome.tabs.update(tabID,{url:page});
  try {
    await account.finish(attempt,url);
    await notice('connected');
  } catch (error) {
    await notice(errorCode(error),diagnosticOf(error));
  }
}

async function cancelSignIn(): Promise<void> {
  const saved = await chrome.storage.session.get('pending');
  const active = pending(saved.pending);
  await chrome.storage.session.remove('pending');
  await notice('idle');
  if (active) await chrome.tabs.update(active.tabID,{url:page});
}

async function dispatch(raw: unknown, sender: chrome.runtime.MessageSender): Promise<unknown> {
  const message = object(raw);
  const owner=sender.documentId ?? 'extension';
  await ready;
  switch (message.type) {
    case 'status': {
      const saved = await chrome.storage.session.get(['pending','notice','diagnostic']);
      const active = pending(saved.pending);
      if (active && active.attempt.expiresAt <= Date.now()) {
        await chrome.storage.session.remove('pending');
        await notice('auth_expired');
        saved.notice = 'auth_expired';
        delete saved.diagnostic;
        delete saved.pending;
      }
      return {...await account.status(),notice:saved.notice ?? 'idle',pending:!!saved.pending,diagnostic:parseDiagnostic(saved.diagnostic)};
    }
    case 'connect': return ordered(start);
    case 'cancelSignIn': return ordered(cancelSignIn);
    case 'logout':
      jobs.cancelAll();
      return ordered(async()=>{await cancelSignIn();await account.logout();});
    case 'models': return models(await account.access());
    case 'settings': return preferences();
    case 'saveSettings': return savePreferences(message.settings);
    case 'currentSite': return currentSite();
    case 'setAutomatic': return ordered(()=>setAutomatic(requiredString(message.origin),message.enabled===true));
    case 'pageAction': {
      const site=await currentSite();
      return sendAction(site.tabID,requiredString(message.action));
    }
    case 'contentConfig': return contentConfig(requiredString(sender.url));
    case 'translateBlock': {
      const config=await preferences();
      return jobs.run(owner,requiredString(message.id),async signal=>translate(await account.access(),requiredString(message.text),config.target,config.model,signal));
    }
    case 'translate': {
      return jobs.run(owner,crypto.randomUUID(),async signal=>translate(await account.access(),requiredString(message.text),requiredString(message.target),requiredString(message.model),signal));
    }
    case 'cancel': jobs.cancel(owner); return;
    default: throw new Error('invalid_input');
  }
}

chrome.webNavigation.onBeforeNavigate.addListener(details=>{
  if (details.frameId !== 0) return;
  void ordered(()=>finish(details.tabId,details.url)).catch(error=>notice(errorCode(error)));
},{url:[{hostEquals:'127.0.0.1',pathEquals:new URL(provider.redirect).pathname}]});
chrome.tabs.onRemoved.addListener(tabID=>{
  for(const [owner,id] of documentTabs)if(id===tabID){jobs.cancel(owner);documentTabs.delete(owner);}
  void ordered(async()=>{
    const saved = await chrome.storage.session.get('pending');
    if (pending(saved.pending)?.tabID !== tabID) return;
    await chrome.storage.session.remove('pending');
    await notice('cancelled');
  });
});
chrome.runtime.onMessage.addListener((message:unknown,sender,respond)=>{
  if (sender.id !== chrome.runtime.id) return false;
  const trustedPage=[page,chrome.runtime.getURL('popup.html')].includes(sender.url ?? '');
  if(!trustedPage) {
    if(sender.tab?.id===undefined || sender.frameId!==0 || !sender.documentId)return false;
    try {
      origin(requiredString(sender.url));
      if(!['translateBlock','contentConfig','cancel'].includes(String(object(message).type)))return false;
    } catch {return false;}
    documentTabs.set(sender.documentId,sender.tab.id);
  }
  void dispatch(message,sender).then(data=>respond({ok:true,data}),async error=>{
    let code=errorCode(error);
    if(code==='sign_in_required') {
      try{await account.logout();}catch(storageError){code=errorCode(storageError);}
    }
    respond({ok:false,error:code,diagnostic:diagnosticOf(error)});
  });
  return true;
});
chrome.webNavigation.onCommitted.addListener(details=>{
  if(details.frameId!==0)return;
  for(const [owner,tabID] of documentTabs)if(tabID===details.tabId && owner!==details.documentId) {
    jobs.cancel(owner);documentTabs.delete(owner);
  }
});
chrome.commands.onCommand.addListener((command,tab)=>{
  if(tab?.id!==undefined)void sendAction(tab.id,command).catch(error=>notice(errorCode(error)));
});
chrome.runtime.onInstalled.addListener(()=>{
  chrome.contextMenus.removeAll(()=>{
    chrome.contextMenus.create({id:'selection',title:chrome.i18n.getMessage('selection'),contexts:['selection']});
    chrome.contextMenus.create({id:'toggle',title:chrome.i18n.getMessage('toggle'),contexts:['page']});
    chrome.contextMenus.create({id:'input',title:chrome.i18n.getMessage('input'),contexts:['editable']});
  });
});
chrome.contextMenus.onClicked.addListener((info,tab)=>{
  if(tab?.id!==undefined)void sendAction(tab.id,String(info.menuItemId)).catch(error=>notice(errorCode(error)));
});
