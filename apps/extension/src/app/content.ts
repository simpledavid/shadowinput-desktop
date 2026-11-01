import {Reader} from '../features/reading/index.js';
import {request} from '../shared/rpc.js';
import {object} from '../shared/validation.js';
import {manualInjectionMarker} from '../shared/config.js';

// An isolated-world marker prevents repeated injection from stacking listeners.
const marker='shadowinput-content-v1';
const manual=Reflect.get(globalThis,manualInjectionMarker)===true;
Reflect.deleteProperty(globalThis,manualInjectionMarker);
if(!Object.prototype.hasOwnProperty.call(globalThis,marker)) {
  Object.defineProperty(globalThis,marker,{value:true});
  const reader=new Reader(document,request);
  chrome.runtime.onMessage.addListener((raw:unknown,sender,respond)=>{
    if(sender.id!==chrome.runtime.id)return false;
    const message=object(raw);
    if(message.type!=='reading' || typeof message.action!=='string')return false;
    respond({ok:true});
    void reader.action(message.action).catch(error=>reader.showError(error));
    return false;
  });
  void reader.initialize(!manual).catch(error=>reader.showError(error));
}
