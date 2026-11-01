import { provider } from '../../shared/config.js';
import { base64url } from '../../shared/validation.js';
import type { Attempt, Registration } from './index.js';

const random = () => base64url(crypto.getRandomValues(new Uint8Array(32)));

export async function begin(record: Registration): Promise<Attempt> {
  const registration = {hostID:record.hostID,clientID:record.clientID,subject:record.subject};
  const state = random(), nonce = random(), verifier = random();
  const challenge = base64url(new Uint8Array(await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))));
  const query = new URLSearchParams({
    client_id: registration.clientID ?? provider.registrationClient,
    ext_agent_host_id: registration.hostID,
    response_type: 'code', redirect_uri: provider.redirect, scope: provider.scope,
    resource: provider.api, state, nonce, code_challenge: challenge, code_challenge_method: 'S256',
  });
  if (!registration.clientID) query.set('agent_name_hint', 'shadowinput');
  return { state, nonce, verifier, registration, redirect: provider.redirect,
    expiresAt: Date.now() + 5 * 60_000,
    authorizationURL: `${provider.issuer}/api/accounts/authorize?${query}` };
}

export function callback(attempt: Attempt, value: string, now = Date.now()): {code: string; clientID: string} {
  const url = new URL(value);
  if (url.origin + url.pathname !== attempt.redirect || url.hash || url.username || url.password || now >= attempt.expiresAt) throw new Error('auth_expired');
  for (const key of ['code', 'state', 'error', 'client_id']) {
    if (url.searchParams.getAll(key).length > 1) throw new Error('invalid_callback');
  }
  if (url.searchParams.get('state') !== attempt.state) throw new Error('invalid_callback');
  if (url.searchParams.has('error')) throw new Error('authorization_denied');
  const code = url.searchParams.get('code');
  const issued = url.searchParams.get('client_id');
  const clientID = attempt.registration.clientID ?? issued;
  if (!code || !clientID || clientID === provider.registrationClient) throw new Error('invalid_callback');
  if (issued && issued !== clientID) throw new Error('invalid_callback');
  return {code, clientID};
}
