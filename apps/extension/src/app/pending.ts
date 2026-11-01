import {parseRecord, type Attempt} from '../features/connect/index.js';
import {provider} from '../shared/config.js';
import {object,requiredString} from '../shared/validation.js';

export function pending(value: unknown): {tabID: number; attempt: Attempt} | undefined {
  if (value === undefined) return undefined;
  const saved=object(value), attempt=object(saved.attempt);
  if (typeof saved.tabID !== 'number' || !Number.isInteger(saved.tabID) ||
    typeof attempt.expiresAt !== 'number' || !Number.isFinite(attempt.expiresAt) || attempt.redirect !== provider.redirect) throw new Error('storage_failed');
  return {tabID:saved.tabID,attempt:{
    state:requiredString(attempt.state),nonce:requiredString(attempt.nonce),verifier:requiredString(attempt.verifier),
    authorizationURL:requiredString(attempt.authorizationURL),redirect:provider.redirect,
    registration:parseRecord(attempt.registration),expiresAt:attempt.expiresAt,
  }};
}
