import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Account} from '../build/features/connect/account.js';
import {signingFixture} from './helpers.mjs';

const initial = () => ({
  hostID:'urn:uuid:00000000-0000-4000-8000-000000000001', clientID:'oaiapp_test', subject:'synthetic-user',
  session:{email:'test@example.invalid', access:'old', refresh:'refresh-old', scope:'chatgpt.tokens.use.direct', expiresAt:1},
});
function storage(record = initial()) {
  return {record, async load(){return structuredClone(this.record);}, async save(next){this.record=structuredClone(next);}};
}

test('parallel requests share one rotating refresh; a restarted owner uses the saved credentials', async () => {
  const store = storage(); let calls=0;
  const fetcher = async (url, init) => {
    calls++;
    assert.equal(new URLSearchParams(init.body).get('refresh_token'), 'refresh-old');
    return Response.json({access_token:'new', refresh_token:'refresh-new', token_type:'Bearer', expires_in:3600});
  };
  const account = new Account(store, fetcher);
  assert.deepEqual(await Promise.all([account.access(),account.access()]), ['new','new']);
  assert.equal(calls,1);
  assert.equal(store.record.session.refresh,'refresh-new');
  assert.equal(await new Account(store,fetcher).access(),'new');
  assert.equal(calls,1);
});

test('revoked plan permission saves rotated credentials but prevents inference', async () => {
  const store = storage();
  const account = new Account(store, async()=>Response.json({access_token:'new',refresh_token:'rotated',token_type:'Bearer',expires_in:3600,scope:'openid'}));
  await assert.rejects(account.access(),/plan_permission/);
  assert.equal(store.record.session.refresh,'rotated');
  await assert.rejects(account.access(),/plan_permission/);
});

test('logout removes credentials and preserves account-bound registration', async () => {
  const store=storage(); const account=new Account(store);
  await account.logout();
  assert.equal(store.record.session,undefined);
  assert.equal(store.record.clientID,'oaiapp_test');
  await assert.rejects(account.access(),/sign_in_required/);
});

test('rejected refresh clears invalid credentials and requests sign-in, without retrying',async()=>{
  const store=storage(); let calls=0;
  const account=new Account(store,async()=>{calls++;return Response.json({error:'invalid_grant'},{status:400});});
  await assert.rejects(account.access(),/sign_in_required/);
  assert.equal(store.record.session,undefined);
  assert.equal(calls,1);
});

test('failed persistence never releases an unpersisted rotated access token', async()=>{
  const store=storage(); store.save=async()=>{throw new Error('storage_failed');};
  const account=new Account(store,async()=>Response.json({access_token:'new',refresh_token:'rotated',token_type:'Bearer',expires_in:3600}));
  await assert.rejects(account.access(),/storage_failed/);
});

test('real signed token exchange connects only after identity and plan permission are validated',async()=>{
  const fixture=await signingFixture();
  const store=storage({hostID:initial().hostID});
  let attempt, scope='openid chatgpt.tokens.use.direct', wrongNonce=false, exchanges=0;
  const account=new Account(store,async(url,init)=>{
    if(url.endsWith('/jwks.json'))return Response.json(fixture.keys);
    exchanges++;
    const form=new URLSearchParams(init.body);
    assert.equal(form.get('client_id'),'oaiapp_test');
    assert.equal(form.get('code_verifier'),attempt.verifier);
    assert.equal(form.get('redirect_uri'),attempt.redirect);
    assert.equal(store.record.clientID,'oaiapp_test');
    return Response.json({token_type:'Bearer',access_token:'synthetic-access',refresh_token:'synthetic-refresh',expires_in:3600,scope,id_token:await fixture.token(attempt,wrongNonce?{nonce:'wrong'}:{})});
  });
  const returnURL=()=>attempt.redirect+'?'+new URLSearchParams({state:attempt.state,code:'synthetic-code',client_id:'oaiapp_test'});
  attempt=await account.start(); wrongNonce=true;
  await assert.rejects(account.finish(attempt,returnURL()),/invalid_identity/);
  assert.equal(store.record.session,undefined);
  attempt=await account.start(); wrongNonce=false;scope='openid';
  await assert.rejects(account.finish(attempt,returnURL()),/plan_permission/);
  assert.equal(store.record.session,undefined);
  attempt=await account.start();scope='openid chatgpt.tokens.use.direct';
  await account.finish(attempt,returnURL());
  assert.deepEqual(await account.status(),{email:'test@example.invalid',connected:true});
  assert.equal(await account.access(),'synthetic-access');
  assert.equal(exchanges,3);
  const next=await account.start();
  assert.equal(next.registration.session,undefined,'pending auth must not duplicate account credentials');
});

test('failed code exchange retains registration so retry does not create another client',async()=>{
  const store=storage({hostID:initial().hostID});
  const account=new Account(store,async()=>Response.json({error:'invalid_grant'},{status:400}));
  const attempt=await account.start();
  await assert.rejects(account.finish(attempt,attempt.redirect+'?'+new URLSearchParams({state:attempt.state,code:'expired',client_id:'oaiapp_test'})),/sign_in_required/);
  assert.equal(new URL((await account.start()).authorizationURL).searchParams.get('client_id'),'oaiapp_test');
  assert.deepEqual(await account.status(),{email:'',connected:false});
});

test('the native fetch receiver remains the worker global when the account owns transport',async()=>{
  const original=globalThis.fetch;
  let requests=0;
  globalThis.fetch=async function() {
    assert.equal(this,globalThis,'browser fetch requires its global receiver');
    requests++;
    return Response.json({error:'invalid_grant'},{status:400});
  };
  try {
    const account=new Account(storage({hostID:initial().hostID}));
    const attempt=await account.start();
    const url=attempt.redirect+'?'+new URLSearchParams({state:attempt.state,code:'synthetic-code',client_id:'oaiapp_test'});
    await assert.rejects(account.finish(attempt,url),/sign_in_required/);
    assert.equal(requests,1,'the request must reach the transport');
  } finally {globalThis.fetch=original;}
});
