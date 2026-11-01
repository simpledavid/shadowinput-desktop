import {mountToolbar} from '../features/toolbar/index.js';
import {copy} from '../shared/copy.js';
import {errorCode} from '../shared/errors.js';
void mountToolbar().catch(error=>{
  const status=document.querySelector('[role="status"]');
  if(status)status.textContent=copy[navigator.language.startsWith('zh')?'zh':'en'].errors[errorCode(error)];
});
