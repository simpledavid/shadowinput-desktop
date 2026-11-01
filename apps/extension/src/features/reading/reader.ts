import {copy} from '../../shared/copy.js';
import {readingCopy} from '../../shared/reading-copy.js';
import {errorCode} from '../../shared/errors.js';
import {settings,type Settings} from '../../shared/settings.js';
import {object,requiredString} from '../../shared/validation.js';
import {paragraphs,originalText,annotate,restore} from './paragraphs.js';
import {draft} from './draft.js';
import {card} from './card.js';

type Send=(type:string,fields?:Record<string,unknown>)=>Promise<unknown>;
export class Reader {
  private config:Settings={target:'zh-CN',model:'',locale:'zh'};
  private enabled=false;
  private generation=0;
  private seen=new WeakSet<Element>();
  private cache=new Map<string,string>();
  private count=0;
  private activeScan:number|undefined;
  private hovered:Element|null=null;
  private hoverEnabled=false;
  private hoverTimer:ReturnType<typeof setTimeout>|undefined;
  private mutationTimer:ReturnType<typeof setTimeout>|undefined;
  private observer:MutationObserver;
  constructor(private document:Document,private send:Send) {
    this.observer=new MutationObserver(changes=>{
      if(!this.enabled || changes.every(change=>Array.from(change.addedNodes).every(node=>node instanceof Element && node.hasAttribute('data-shadowinput'))))return;
      clearTimeout(this.mutationTimer);
      this.mutationTimer=setTimeout(()=>void this.scan(),400);
    });
    this.observer.observe(document.documentElement,{childList:true,subtree:true});
    document.addEventListener('pointerover',event=>this.hover(event));
  }
  async initialize(automatic=true):Promise<void> {
    const config=object(await this.send('contentConfig'));
    this.config=settings(config);
    if(automatic && config.automatic===true)await this.action('toggle');
  }
  private get text(){return readingCopy[this.config.locale];}
  private error(error:unknown):string{return copy[this.config.locale].errors[errorCode(error)];}
  showError(error:unknown):void{this.panel('shadowinput').body.textContent=this.error(error);}
  private panel(title:string) {
    return card(this.document,this.config.locale,title,()=>{});
  }
  private async result(text:string):Promise<string> {
    const generation=this.generation;
    const key=this.config.model+'|'+this.config.target+'|'+text;
    const cached=this.cache.get(key);
    if(cached)return cached;
    const translated=requiredString(await this.send('translateBlock',{text,id:crypto.randomUUID()}));
    if(generation!==this.generation)throw new Error('cancelled');
    if(this.cache.size>=200)this.cache.delete(this.cache.keys().next().value ?? '');
    this.cache.set(key,translated);return translated;
  }
  async action(action:string):Promise<void> {
    const started=this.generation;
    try {
      this.config=settings(await this.send('contentConfig'));
      if(!this.config.model)throw new Error('sign_in_required');
      if(action==='toggle') {
        if(this.enabled || this.document.querySelector('[data-shadowinput="translation"]')){this.stop();return;}
        this.enabled=true;this.count=0;this.seen=new WeakSet();this.generation++;
        await this.scan();
      } else if(action==='selection')await this.selection();
      else if(action==='input')await this.input();
      else if(action==='hover') {
        this.hoverEnabled=!this.hoverEnabled;
        this.panel('shadowinput').body.textContent=this.hoverEnabled?this.text.hoverReady:this.text.close;
      }
    } catch(error){if(started===this.generation)this.panel('shadowinput').body.textContent=this.error(error);}
  }
  private stop():void {
    this.enabled=false;this.generation++;this.activeScan=undefined;this.observer.takeRecords();
    restore(this.document);this.cache.clear();this.seen=new WeakSet();
    void this.send('cancel').catch(error=>{this.panel('shadowinput').body.textContent=this.error(error);});
    this.document.querySelector('[data-shadowinput="panel"]')?.remove();
  }
  private async scan():Promise<void> {
    if(this.activeScan===this.generation || !this.enabled || this.count>=120)return;
    const generation=this.generation;
    this.activeScan=generation;
    const nodes=paragraphs(this.document).filter(node=>!this.seen.has(node)).slice(0,120-this.count);
    if(!nodes.length){this.activeScan=undefined;return;}
    const panel=card(this.document,this.config.locale,this.text.pageTitle,()=>this.stop());
    panel.button(this.text.restore,()=>this.stop());
    panel.body.textContent=this.text.working;
    let cursor=0;
    const work=async()=>{
      while(cursor<nodes.length && this.enabled && generation===this.generation) {
        const node=nodes[cursor++];if(!node)break;
        this.seen.add(node);
        const before=originalText(node);
        const translated=await this.result(before);
        if(!this.enabled || generation!==this.generation)return;
        if(annotate(node,before,translated))this.count++;
        panel.body.textContent=this.count+' '+this.text.progress+(this.count>=120?' · '+this.text.limited:'');
      }
    };
    try{await Promise.all([work(),work()]);}
    catch(error){
      if(generation===this.generation) {
        panel.body.textContent=this.error(error);this.enabled=false;this.generation++;await this.send('cancel');
      }
    }
    finally{if(this.activeScan===generation)this.activeScan=undefined;}
  }
  private async selection():Promise<void> {
    const selected=this.document.getSelection()?.toString().trim() ?? '';
    const panel=this.panel(this.text.selectionTitle);
    if(!selected){panel.body.textContent=this.text.noSelection;return;}
    panel.body.textContent=this.text.working;
    const result=await this.result(selected);
    if(!panel.host.isConnected)return;
    panel.body.textContent=result;
    const button=panel.button(this.text.copy,()=>void navigator.clipboard.writeText(result).then(()=>{button.textContent=this.text.copied;},error=>{panel.body.textContent=this.error(error);}));
  }
  private async input():Promise<void> {
    const pending=draft(this.document.activeElement);
    const panel=this.panel(this.text.inputTitle);
    if(!pending){panel.body.textContent=this.text.noInput;return;}
    panel.body.textContent=this.text.working;
    const result=await this.result(pending.text);
    if(!panel.host.isConnected)return;
    panel.body.textContent=result;
    const apply=panel.button(this.text.apply,()=>{
      if(pending.apply(result))panel.host.remove();
      else {panel.body.textContent=copy[this.config.locale].errors.draft_changed;apply.disabled=true;}
    });
  }
  private hover(event:PointerEvent):void {
    clearTimeout(this.hoverTimer);
    if(!this.hoverEnabled || !event.shiftKey || !(event.target instanceof Element))return;
    const node=event.target.closest('p,li,blockquote,h1,h2,h3,td');
    if(!node || node.closest('[data-shadowinput]') || node===this.hovered)return;
    this.hovered=node;
    this.hoverTimer=setTimeout(()=>{
      const generation=this.generation;
      const text=originalText(node);
      if(!text)return;
      void this.result(text).then(result=>{if(generation===this.generation)annotate(node,text,result);},error=>{this.panel('shadowinput').body.textContent=this.error(error);});
    },300);
  }
}
