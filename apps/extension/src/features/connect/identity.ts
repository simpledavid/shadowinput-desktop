import { provider } from '../../shared/config.js';
import { object, requiredString, unbase64url } from '../../shared/validation.js';
import type { Attempt } from './index.js';

function jsonPart(value: string): Record<string, unknown> {
  return object(JSON.parse(new TextDecoder().decode(unbase64url(value))));
}

function publicKey(keys: unknown, kid: string): JsonWebKey {
  const set = object(keys).keys;
  if (!Array.isArray(set) || set.length > 100) throw new Error('invalid_identity');
  const matches = set.map(object).filter(k => k.kid === kid && k.kty === 'RSA');
  const key = matches[0];
  if (matches.length !== 1 || !key || (key.use && key.use !== 'sig') || (key.alg && key.alg !== 'RS256')) throw new Error('invalid_identity');
  return {kty: 'RSA', n: requiredString(key.n), e: requiredString(key.e), alg: 'RS256', ext: true};
}

function validateClaims(claims: Record<string, unknown>, attempt: Attempt, clientID: string): void {
  const now = Date.now() / 1000;
  const aud = claims.aud;
  const audienceMatches = aud === clientID || (Array.isArray(aud) && aud.includes(clientID) && (aud.length === 1 || claims.azp === clientID));
  if (claims.iss !== provider.issuer || !audienceMatches || claims.nonce !== attempt.nonce) throw new Error('invalid_identity');
  if (typeof claims.exp !== 'number' || claims.exp <= now || (claims.nbf !== undefined && (typeof claims.nbf !== 'number' || claims.nbf > now))) throw new Error('invalid_identity');
  if (attempt.registration.subject && claims.sub !== attempt.registration.subject) throw new Error('account_mismatch');
}

export async function verifyIdentity(token: string, keys: unknown, attempt: Attempt, clientID: string): Promise<{subject: string; email: string}> {
  if (token.length > 64_000) throw new Error('invalid_identity');
  const parts = token.split('.');
  const [head, body, signature] = parts;
  if (parts.length !== 3 || !head || !body || !signature) throw new Error('invalid_identity');
  const header = jsonPart(head);
  if (header.alg !== 'RS256' || header.crit !== undefined) throw new Error('invalid_identity');
  const key = await crypto.subtle.importKey('jwk', publicKey(keys, requiredString(header.kid)), {name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256'}, false, ['verify']);
  const valid = await crypto.subtle.verify('RSASSA-PKCS1-v1_5', key, unbase64url(signature), new TextEncoder().encode(`${head}.${body}`));
  if (!valid) throw new Error('invalid_identity');
  const claims = jsonPart(body);
  validateClaims(claims, attempt, clientID);
  return {subject: requiredString(claims.sub), email: typeof claims.email === 'string' ? claims.email : ''};
}
