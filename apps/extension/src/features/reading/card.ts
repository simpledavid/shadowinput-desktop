import {readingCopy} from '../../shared/reading-copy.js';
import type {Locale} from '../../shared/copy.js';

export function card(document:Document,locale:Locale,title:string,onClose:()=>void) {
  const t=readingCopy[locale];
  document.querySelector('[data-shadowinput="panel"]')?.remove();
  const host=document.createElement('div');
  host.dataset.shadowinput='panel';
  host.style.cssText='all:initial;position:fixed;z-index:2147483647;right:20px;bottom:20px;width:min(360px,calc(100vw - 40px));';
  const root=host.attachShadow({mode:'closed'});
  const style=document.createElement('style');
  style.textContent=':host{color-scheme:light dark}section{font:14px/1.6 system-ui;color:#25221e;background:#fffdf8;border:1px solid #e7e1d8;border-radius:14px;padding:16px;box-shadow:0 8px 30px #0002}header,footer{display:flex;gap:10px;align-items:center}header{justify-content:space-between;font-weight:650}p{white-space:pre-wrap;max-height:35vh;overflow:auto;overflow-wrap:anywhere;margin:14px 0}button{font:inherit;color:inherit;background:transparent;border:1px solid #cec7b9;border-radius:7px;min-height:32px;padding:4px 10px;cursor:pointer}button:focus-visible{outline:2px solid #0866ff}button:disabled{opacity:.5;cursor:default}footer:empty{display:none}@media(prefers-color-scheme:dark){section{color:#f7f4ee;background:#211e1a;border-color:#4e463a}button{border-color:#62594c}}';
  const section=document.createElement('section'),header=document.createElement('header'),label=document.createElement('span');
  label.textContent=title;
  const close=document.createElement('button');close.type='button';close.textContent=t.close;
  close.addEventListener('click',()=>{host.remove();onClose();});
  header.append(label,close);
  const body=document.createElement('p');body.setAttribute('role','status');
  const footer=document.createElement('footer');
  section.append(header,body,footer);root.append(style,section);document.documentElement.append(host);
  return {body,host,button(text:string,action:()=>void) {
    const button=document.createElement('button');button.textContent=text;button.type='button';
    button.addEventListener('click',action);footer.append(button);return button;
  }};
}
