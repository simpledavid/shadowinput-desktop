const excluded = 'nav,header,footer,script,style,pre,code,noscript,svg,math,button,input,textarea,select,[contenteditable]:not([contenteditable="false"]),[hidden],[aria-hidden="true"],[data-shadowinput],[role="navigation"]';
const blocks = 'p,h1,h2,h3,h4,h5,h6,li,blockquote,td,th,figcaption,div';

function ignored(element: Element): boolean {
  if (element.closest(excluded)) return true;
  const view=element.ownerDocument.defaultView;
  for(let node:Element|null=element;node;node=node.parentElement) {
    const style=view?.getComputedStyle(node);
    if(style?.display==='none' || style?.visibility==='hidden')return true;
  }
  return false;
}

export function originalText(element: Element): string {
  const walker=element.ownerDocument.createTreeWalker(element,4);
  let text='',node:Node|null;
  while((node=walker.nextNode())) {
    if(node.parentElement && !ignored(node.parentElement))text+=node.textContent ?? '';
  }
  return text.trim();
}

export function paragraphs(document: Document): Element[] {
  const root=document.querySelector('main,article,[role="main"]') ?? document.body;
  if(!root)return [];
  return Array.from(root.querySelectorAll(blocks)).filter(node=>{
    if(ignored(node) || Array.from(node.querySelectorAll(blocks)).some(child=>!ignored(child)))return false;
    const text=originalText(node);
    return text.length>=3 && text.length<=8000 && /\p{L}/u.test(text);
  });
}

export function annotate(element: Element, before: string, translation: string): boolean {
  if(!element.isConnected || originalText(element)!==before || !translation.trim())return false;
  if(element.querySelector(':scope > [data-shadowinput]'))return false;
  const host=element.ownerDocument.createElement('span');
  host.dataset.shadowinput='translation';
  host.style.cssText='display:block!important;margin-block:0.55em!important;white-space:pre-wrap!important;';
  const shadow=host.attachShadow({mode:'open'});
  const style=element.ownerDocument.createElement('style');
  style.textContent=':host{color:inherit;font:inherit}span{display:block;border-inline-start:2px solid #b7a47b;padding-inline-start:.7em;line-height:1.7;overflow-wrap:anywhere}';
  const content=element.ownerDocument.createElement('span');
  content.textContent=translation;
  shadow.append(style,content);
  element.append(host);
  return true;
}

export function restore(document: Document): void {
  document.querySelectorAll('[data-shadowinput="translation"]').forEach(node=>node.remove());
}
