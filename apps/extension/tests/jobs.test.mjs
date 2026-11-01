import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Jobs} from '../build/features/translate/jobs.js';
test('cancellation belongs to a document and a cancelled response cannot return late',async()=>{
  const jobs=new Jobs();let releaseA,releaseB;
  const a=jobs.run('document-a','a',()=>new Promise(resolve=>{releaseA=resolve;}));
  const b=jobs.run('document-b','b',()=>new Promise(resolve=>{releaseB=resolve;}));
  jobs.cancel('document-a');releaseA('stale');releaseB('current');
  await assert.rejects(a,/cancelled/);assert.equal(await b,'current');
});
test('duplicate jobs and excess document work do not multiply provider requests',async()=>{
  const jobs=new Jobs();const finishes=[];let calls=0;
  const work=()=>{calls++;return new Promise(resolve=>finishes.push(resolve));};
  const a=jobs.run('document','one',work),b=jobs.run('document','two',work);
  await assert.rejects(jobs.run('document','one',work),/busy/);
  await assert.rejects(jobs.run('document','three',work),/busy/);
  assert.equal(calls,2);
  finishes.forEach(done=>done('ok'));await Promise.all([a,b]);
  assert.equal(await jobs.run('document','one',async()=> 'again'),'again');
});
