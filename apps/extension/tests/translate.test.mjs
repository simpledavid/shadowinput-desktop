import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readTranslation,translate,models} from '../build/features/translate/index.js';

function stream(events, split = 9) {
  const bytes=new TextEncoder().encode(events.map(e=>`data: ${JSON.stringify(e)}\r\n\r\n`).join(''));
  return new Response(new ReadableStream({start(controller){
    for(let n=0;n<bytes.length;n+=split)controller.enqueue(bytes.slice(n,n+split));
    controller.close();
  }}));
}
test('translation handles split UTF-8 and only returns after provider completion',async()=>{
  const result=await readTranslation(stream([
    {type:'response.output_text.delta',delta:'你好，'},
    {type:'response.output_text.delta',delta:'世界。'},
    {type:'response.completed',response:{status:'completed'}},
  ]));
  assert.equal(result,'你好，世界。');
});
test('truncated, refused, failed and incomplete text can never be applied',async()=>{
  const delta={type:'response.output_text.delta',delta:'partial'};
  for(const tail of [[],[{type:'response.incomplete'}],[{type:'response.refusal.delta'}],[{type:'response.failed'}],[{type:'response.completed',response:{status:'incomplete'}}]]) {
    await assert.rejects(readTranslation(stream([delta,...tail])));
  }
});
test('oversized and malformed provider output is rejected',async()=>{
  await assert.rejects(readTranslation(stream([{type:'response.output_text.delta',delta:'a'.repeat(256001)},{type:'response.completed',response:{status:'completed'}}], 8192)));
  await assert.rejects(readTranslation(new Response('data: invalid\n\n')));
});

test('translation submits only the requested text using the documented plan request shape',async()=>{
  let calls=0;
  const fetcher=async(url,options)=>{
    calls++;
    assert.equal(url,'https://api.openai.com/v1/responses');
    assert.equal(options.headers.Authorization,'Bearer synthetic');
    assert.equal(options.credentials,'omit');
    assert.equal(options.redirect,'error');
    const body=JSON.parse(options.body);
    assert.deepEqual(body.input,[{role:'user',content:'Hello'}]);
    assert.equal(body.store,false);assert.equal(body.stream,true);
    assert.equal(body.temperature,undefined);assert.equal(body.max_output_tokens,undefined);
    return stream([{type:'response.output_text.delta',delta:'你好'},{type:'response.completed',response:{status:'completed'}}]);
  };
  const signal=new AbortController().signal;
  assert.equal(await translate('synthetic','Hello','zh-CN','model-test',signal,fetcher),'你好');
  for(const args of [['','zh-CN','model-test'],['a'.repeat(8001),'en','model-test'],['Hello','invalid','model-test'],['Hello','en','']]) {
    await assert.rejects(translate('synthetic',...args,signal,fetcher),/invalid_input/);
  }
  assert.equal(calls,1);
});

test('model picker uses only the authorized visible catalog',async()=>{
  const list=await models('synthetic',async()=>Response.json({models:[{slug:'allowed',display_name:'Allowed',visibility:'list'},{slug:'hidden',display_name:'Hidden',visibility:'hidden'}]}));
  assert.deepEqual(list,[{id:'allowed',name:'Allowed'}]);
  await assert.rejects(models('synthetic',async()=>Response.json({models:[]})),/plan_permission/);
  await assert.rejects(models('synthetic',async()=>new Response('',{status:401})),/sign_in_required/);
});
