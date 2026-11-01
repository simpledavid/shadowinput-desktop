import { test } from 'node:test';
import assert from 'node:assert/strict';
import { begin, callback, verifyIdentity } from '../build/features/connect/index.js';

const hostID = 'urn:uuid:00000000-0000-4000-8000-000000000001';
const encode = value => Buffer.from(JSON.stringify(value)).toString('base64url');
const callbackURL = (attempt, extra = {}) => `${attempt.redirect}?${new URLSearchParams({
  code: 'synthetic-code', state: attempt.state, client_id: 'oaiapp_test', ...extra,
})}`;

test('new installation requests plan permission with fresh PKCE and one stable host identity', async () => {
  const a = await begin({ hostID });
  const b = await begin({ hostID });
  const url = new URL(a.authorizationURL);
  assert.equal(url.origin, 'https://auth.openai.com');
  assert.equal(url.searchParams.get('ext_agent_host_id'), hostID);
  assert.equal(url.searchParams.get('client_id'), 'dynamic_agent_client');
  assert.equal(url.searchParams.get('agent_name_hint'), 'shadowinput');
  assert.ok(url.searchParams.get('scope').split(' ').includes('chatgpt.tokens.use.direct'));
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(a.verifier));
  assert.equal(url.searchParams.get('code_challenge'), Buffer.from(digest).toString('base64url'));
  assert.notEqual(a.state, b.state);
  assert.notEqual(a.nonce, b.nonce);
  assert.equal(callback(a, callbackURL(a)).clientID, 'oaiapp_test');
});

test('callback rejects forged, expired, ambiguous and denied authorizations', async () => {
  const a = await begin({hostID});
  for (const url of [
    callbackURL(a, {state: 'wrong'}), callbackURL(a, {code: ''}),
    callbackURL(a, {client_id: 'dynamic_agent_client'}),
    callbackURL(a, {error: 'access_denied'}),
    callbackURL(a).replace('/auth/callback', '/other'),
    callbackURL(a).replace('127.0.0.1', 'localhost'),
    callbackURL(a) + '&state=another',
  ]) assert.throws(() => callback(a, url));
  assert.throws(() => callback(a, callbackURL(a), a.expiresAt + 1));
});

test('returning account retains its exact issued registration', async () => {
  const a = await begin({hostID, clientID: 'oaiapp_old', subject: 'synthetic-user'});
  const url = new URL(a.authorizationURL);
  assert.equal(url.searchParams.get('client_id'), 'oaiapp_old');
  assert.equal(url.searchParams.has('agent_name_hint'), false);
  const cb = new URL(callbackURL(a)); cb.searchParams.delete('client_id');
  assert.equal(callback(a, cb.href).clientID, 'oaiapp_old');
  assert.throws(() => callback(a, callbackURL(a)));
});

test('identity requires a valid signature, issuer, audience, nonce, expiry and account binding', async () => {
  const a = await begin({hostID, subject: 'synthetic-user'});
  const pair = await crypto.subtle.generateKey({name: 'RSASSA-PKCS1-v1_5', modulusLength: 2048, publicExponent: new Uint8Array([1,0,1]), hash: 'SHA-256'}, true, ['sign','verify']);
  const jwk = {...await crypto.subtle.exportKey('jwk', pair.publicKey), kid: 'test', alg: 'RS256', use: 'sig'};
  const claims = {iss: 'https://auth.openai.com', aud: 'oaiapp_test', sub: 'synthetic-user', nonce: a.nonce, exp: Math.floor(Date.now()/1000)+600, email: 'test@example.invalid'};
  async function signed(overrides = {}, header = {}) {
    const content = `${encode({alg:'RS256',kid:'test',...header})}.${encode({...claims,...overrides})}`;
    const signature = await crypto.subtle.sign('RSASSA-PKCS1-v1_5', pair.privateKey, new TextEncoder().encode(content));
    return `${content}.${Buffer.from(signature).toString('base64url')}`;
  }
  assert.deepEqual(await verifyIdentity(await signed(), {keys:[jwk]}, a, 'oaiapp_test'), {subject:'synthetic-user',email:'test@example.invalid'});
  for (const overrides of [{iss:'https://evil.invalid'}, {aud:'wrong'}, {nonce:'wrong'}, {exp:1}, {sub:'other-user'}, {nbf:Math.floor(Date.now()/1000)+600}]) {
    await assert.rejects(verifyIdentity(await signed(overrides), {keys:[jwk]}, a, 'oaiapp_test'));
  }
  await assert.rejects(verifyIdentity(await signed({}, {alg:'none'}), {keys:[jwk]}, a, 'oaiapp_test'));
  const token = await signed();
  const parts = token.split('.'); parts[1] = encode({...claims, email:'forged@example.invalid'});
  await assert.rejects(verifyIdentity(parts.join('.'), {keys:[jwk]}, a, 'oaiapp_test'));
});
