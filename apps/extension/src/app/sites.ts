import {settings,type Settings} from '../shared/settings.js';
import {object} from '../shared/validation.js';
import {manualInjectionMarker} from '../shared/config.js';

export function origin(value:string):string {
  const url=new URL(value);
  if(!['http:','https:'].includes(url.protocol) || ['auth.openai.com','api.openai.com'].includes(url.hostname))throw new Error('unsupported_page');
  return url.origin;
}

export async function preferences():Promise<Settings> {
  const saved=await chrome.storage.local.get('settings');
  return settings(saved.settings);
}
export async function savePreferences(value:unknown):Promise<void> {
  await chrome.storage.local.set({settings:settings(value)});
}

export async function currentSite():Promise<{tabID:number;origin:string;automatic:boolean}> {
  const [tab]=await chrome.tabs.query({active:true,currentWindow:true});
  if(tab?.id===undefined || !tab.url)throw new Error('unsupported_page');
  const host=origin(tab.url), saved=await chrome.storage.local.get('sites');
  const sites=saved.sites===undefined?{}:object(saved.sites);
  return {tabID:tab.id,origin:host,automatic:sites[host]===true && await chrome.permissions.contains({origins:[host+'/*']})};
}

export async function setAutomatic(host:string,enabled:boolean):Promise<void> {
  const normalized=origin(host);
  if(normalized!==host)throw new Error('invalid_input');
  const bytes=new Uint8Array(await crypto.subtle.digest('SHA-256',new TextEncoder().encode(host)));
  const id='site-'+Array.from(bytes).map(n=>n.toString(16).padStart(2,'0')).join('');
  if(enabled && !await chrome.permissions.contains({origins:[host+'/*']}))throw new Error('site_permission');
  const registered=await chrome.scripting.getRegisteredContentScripts({ids:[id]});
  if(registered.length)await chrome.scripting.unregisterContentScripts({ids:[id]});
  if(enabled)await chrome.scripting.registerContentScripts([{id,matches:[host+'/*'],js:['content.js'],runAt:'document_idle',persistAcrossSessions:true}]);
  const saved=await chrome.storage.local.get('sites');
  const sites=saved.sites===undefined?{}:object(saved.sites);
  if(enabled)sites[host]=true;else delete sites[host];
  await chrome.storage.local.set({sites});
}

export async function contentConfig(url:string):Promise<Settings&{automatic:boolean}> {
  const host=origin(url),saved=await chrome.storage.local.get('sites');
  const sites=saved.sites===undefined?{}:object(saved.sites);
  return {...await preferences(),automatic:sites[host]===true};
}

export async function sendAction(tabID:number,action:string):Promise<void> {
  const tab=await chrome.tabs.get(tabID);
  origin(tab.url ?? '');
  await chrome.scripting.executeScript({target:{tabId:tabID},func:(key:string)=>{Reflect.set(globalThis,key,true);},args:[manualInjectionMarker]});
  await chrome.scripting.executeScript({target:{tabId:tabID},files:['content.js']});
  await chrome.tabs.sendMessage(tabID,{type:'reading',action});
}
