import {object} from './validation.js';
export type Settings = {target:'zh-CN'|'en'|'ja';model:string;locale:'zh'|'en'};
export function settings(value:unknown):Settings {
  if(value===undefined)return {target:'zh-CN',model:'',locale:'zh'};
  const data=object(value);
  if((data.target!=='zh-CN' && data.target!=='en' && data.target!=='ja') || typeof data.model!=='string' || (data.locale!=='zh' && data.locale!=='en'))throw new Error('invalid_input');
  return {target:data.target,model:data.model,locale:data.locale};
}
