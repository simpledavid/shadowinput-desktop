import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Account} from '../build/features/connect/account.js';
import {models} from '../build/features/translate/request.js';
import {request} from '../build/shared/rpc.js';
import {diagnosticOf,diagnosticText} from '../build/shared/diagnostics.js';
import {copy} from '../build/shared/copy.js';

function store() {
  return {
    value:{hostID:'urn:uuid:00000000-0000-4000-8000-000000000001'},
    async load(){return structuredClone(this.value);},
    async save(value){this.value=structuredClone(value);},
  };
}
async function finish(fetcher) {
  const account=new Account(store(),fetcher),attempt=await account.start();
  const url=attempt.redirect+'?'+new URLSearchParams({state:attempt.state,code:'private-code-canary',client_id:'oaiapp_test'});
  return account.finish(attempt,url);
}

test('authorization transport failures identify the failed step without exposing request or error text',async()=>{
  await assert.rejects(finish(async()=>{throw new TypeError('private-code-canary at https://secret.invalid/');}),error=>{
    assert.deepEqual(error.diagnostic,{stage:'token_exchange',code:'network_error',kind:'transport'});
    assert.doesNotMatch(JSON.stringify(error),/private-code-canary|secret\.invalid/);
    return true;
  });
});

test('a key endpoint HTTP failure remains distinct from token exchange failure',async()=>{
  await assert.rejects(finish(async url=>url.endsWith('/jwks.json')?
    new Response('private-response-canary',{status:502}):Response.json({id_token:'private-token-canary'})),error=>{
    assert.deepEqual(error.diagnostic,{stage:'identity_keys',code:'network_error',kind:'http',status:502});
    assert.doesNotMatch(JSON.stringify(error),/private-(response|token)-canary/);
    return true;
  });
});

test('model request timeouts identify model discovery instead of blaming sign-in',async()=>{
  await assert.rejects(models('private-access-canary',async()=>{throw new DOMException('private-message','TimeoutError');}),error=>{
    assert.deepEqual(error.diagnostic,{stage:'models',code:'network_error',kind:'timeout'});
    assert.doesNotMatch(JSON.stringify(error),/private/);
    return true;
  });
});

test('workbench errors survive Chrome message serialization with only allowlisted details',async()=>{
  globalThis.chrome={runtime:{async sendMessage(){return JSON.parse(JSON.stringify({
    ok:false,error:'network_error',diagnostic:{stage:'token_exchange',code:'network_error',kind:'http',status:503,body:'private-body',url:'private-url'},
  }));}}};
  await assert.rejects(request('models'),error=>{
    assert.equal(diagnosticText(diagnosticOf(error),copy.zh.stages),' [授权码兑换 · http · HTTP 503]');
    assert.doesNotMatch(JSON.stringify(error),/private/);
    return true;
  });
});
