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

test('a completed response supplies authoritative text even without text deltas',async()=>{
  const completed={type:'response.completed',response:{status:'completed',output:[
    {type:'message',role:'assistant',status:'completed',content:[{type:'output_text',text:'完整译文。'}]},
  ]}};
  assert.equal(await readTranslation(stream([completed])),'完整译文。');
  assert.equal(await readTranslation(stream([{type:'response.output_text.delta',delta:'不完整'},completed])),'完整译文。');
  const refused=structuredClone(completed);
  refused.response.output[0].content=[{type:'refusal',refusal:'private-refusal-canary'}];
  await assert.rejects(readTranslation(stream([refused])),/translation_refused/);
});

test('stream failures distinguish plan usage, availability, permission and model errors',async()=>{
  const cases=[
    ['subscription_sharing_usage_limit_exceeded','usage_limit'],
    ['rate_limit_exceeded','usage_limit'],
    ['subscription_sharing_usage_unavailable','service_unavailable'],
    ['subscription_sharing_user_unavailable','service_unavailable'],
    ['server_error','service_unavailable'],
    ['subscription_sharing_user_not_eligible','plan_restricted'],
    ['subscription_sharing_unsupported_capability','request_rejected'],
    ['model_not_found','model_unavailable'],
    ['model_not_available','model_unavailable'],
  ];
  for(const [code,expected] of cases) {
    for(const event of [{type:'response.failed',response:{error:{code,message:'private-message-canary'}}},{type:'error',code,message:'private-message-canary'}]) {
      await assert.rejects(translate('synthetic','Hello','zh-CN','model-test',new AbortController().signal,async()=>stream([event])),error=>{
        assert.equal(error.message,expected);
        assert.equal(error.diagnostic.reason,code);
        assert.doesNotMatch(JSON.stringify(error),/private-message-canary/);
        return true;
      });
    }
  }
});

test('interrupted, limited and unknown streams expose safe distinct reasons',async()=>{
  const cases=[
    [[{type:'response.output_text.delta',delta:'partial'}],'incomplete_response','stream_interrupted'],
    [[{type:'response.incomplete',response:{incomplete_details:{reason:'max_output_tokens'}}}],'output_limit','max_output_tokens'],
    [[{type:'response.refusal.done',refusal:'private-refusal-canary'}],'translation_refused','refusal'],
    [[{type:'response.failed',response:{error:{code:'private-code-canary',message:'private-message-canary'}}}],'incomplete_response','provider_failed'],
  ];
  for(const [events,code,reason] of cases) {
    await assert.rejects(translate('synthetic','Hello','zh-CN','model-test',new AbortController().signal,async()=>stream(events)),error=>{
      assert.equal(error.message,code);
      assert.equal(error.diagnostic.reason,reason);
      assert.doesNotMatch(JSON.stringify(error),/private-|partial/);
      return true;
    });
  }
});

test('final output never turns partial, refused or empty assistant content into a translation',async()=>{
  const partial={type:'response.output_text.delta',delta:'partial-canary'};
  const message={type:'message',role:'assistant',status:'completed',content:[{type:'output_text',text:'Final translation'}]};
  for(const output of [[],[{...message,status:'incomplete'}],[{...message,content:[{type:'refusal'}]}],[{type:'reasoning',summary:[]}]]) {
    await assert.rejects(readTranslation(stream([partial,{type:'response.completed',response:{status:'completed',output}}])));
  }
  const output=[{...message,phase:'commentary',content:[{type:'output_text',text:'private-commentary-canary'}]},message];
  assert.equal(await readTranslation(stream([partial,{type:'response.completed',response:{status:'completed',output}}])),'Final translation');
});
