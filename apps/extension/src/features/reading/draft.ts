type Draft = {text: string; apply:(text:string)=>boolean};
export function draft(element: Element | null): Draft | undefined {
  if(!element || !element.isConnected)return;
  const view=element.ownerDocument.defaultView;
  if(!view)return;
  let read:()=>string,write:(text:string)=>void;
  if(element instanceof view.HTMLTextAreaElement ||
    (element instanceof view.HTMLInputElement && ['text','search'].includes(element.type))) {
    if(element.disabled || element.readOnly)return;
    const field=element;
    read=()=>field.value;
    const prototype=field instanceof view.HTMLTextAreaElement ? view.HTMLTextAreaElement.prototype : view.HTMLInputElement.prototype;
    const setter=Object.getOwnPropertyDescriptor(prototype,'value')?.set;
    if(!setter)return;
    write=text=>setter.call(field,text);
  } else if(element instanceof view.HTMLElement && ['true','plaintext-only',''].includes(element.getAttribute('contenteditable') ?? 'false') && element.childElementCount===0) {
    read=()=>element.textContent ?? '';
    write=text=>{element.textContent=text;};
  } else return;
  const text=read();
  if(!text.trim() || text.length>8000)return;
  return {text,apply(next) {
    if(!element.isConnected || read()!==text)return false;
    write(next);
    element.dispatchEvent(new view.InputEvent('input',{bubbles:true,inputType:'insertReplacementText',data:next}));
    return true;
  }};
}
