export type Registration = { hostID: string; clientID?: string; subject?: string };
export type Attempt = {
  state: string; nonce: string; verifier: string; redirect: string;
  expiresAt: number; registration: Registration; authorizationURL: string;
};
export { begin, callback } from './oauth.js';
export { verifyIdentity } from './identity.js';
export { Account, parseRecord, type AccountRecord, type AccountStore } from './account.js';
