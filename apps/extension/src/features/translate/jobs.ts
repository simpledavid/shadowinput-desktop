export class Jobs {
  private active=new Map<string,{owner:string;controller:AbortController}>();
  async run<T>(owner: string, id: string, action:(signal:AbortSignal)=>Promise<T>): Promise<T> {
    const key=owner+':'+id;
    const live=[...this.active.values()].filter(job=>!job.controller.signal.aborted);
    if(this.active.has(key) || live.length>=4 || live.filter(job=>job.owner===owner).length>=2)throw new Error('busy');
    const controller=new AbortController();
    this.active.set(key,{owner,controller});
    try {
      const result=await action(controller.signal);
      if(controller.signal.aborted)throw new Error('cancelled');
      return result;
    } finally {this.active.delete(key);}
  }
  cancel(owner:string): void {
    for(const job of this.active.values())if(job.owner===owner)job.controller.abort();
  }
  cancelAll(): void {for(const job of this.active.values())job.controller.abort();}
}
