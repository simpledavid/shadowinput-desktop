import { provider } from '../../shared/config.js';
import { checked, json, networkOptions, type Fetcher } from '../../shared/http.js';
import { object, requiredString } from '../../shared/validation.js';
import { begin, callback } from './oauth.js';
import { verifyIdentity } from './identity.js';
import type { Attempt, Registration } from './index.js';
import {atStage} from '../../shared/diagnostics.js';
import {HttpError} from '../../shared/errors.js';

export type Session = {email: string; access: string; refresh: string; scope: string; expiresAt: number};
export type AccountRecord = Registration & {session?: Session};
export interface AccountStore {
  load(): Promise<AccountRecord>;
  save(value: AccountRecord): Promise<void>;
}

export function parseRecord(value: unknown): AccountRecord {
  const data = object(value);
  const record: AccountRecord = {hostID: requiredString(data.hostID)};
  if (data.clientID !== undefined) record.clientID = requiredString(data.clientID);
  if (data.subject !== undefined) record.subject = requiredString(data.subject);
  if (data.session !== undefined) {
    const s = object(data.session);
    if (!record.clientID || !record.subject || typeof s.email !== 'string' || typeof s.expiresAt !== 'number' || !Number.isFinite(s.expiresAt)) throw new Error('storage_failed');
    record.session = {email:s.email,access:requiredString(s.access),refresh:requiredString(s.refresh),scope:requiredString(s.scope),expiresAt:s.expiresAt};
  }
  return record;
}

function tokens(data: Record<string, unknown>, previous?: Session): Omit<Session, 'email'> {
  if (typeof data.expires_in !== 'number' || data.expires_in <= 0 || !Number.isFinite(data.expires_in) || data.token_type !== 'Bearer') throw new Error('invalid_response');
  return {
    access: requiredString(data.access_token), refresh: requiredString(data.refresh_token ?? previous?.refresh),
    scope: requiredString(data.scope ?? previous?.scope), expiresAt: Date.now() + data.expires_in * 1000,
  };
}

function requirePlan(session: Session): void {
  if (!session.scope.split(' ').includes(provider.planScope)) throw new Error('plan_permission');
}

export class Account {
  private tail: Promise<unknown> = Promise.resolve();
  constructor(private store: AccountStore, private fetcher: Fetcher = fetch) {}

  private exclusive<T>(operation: () => Promise<T>): Promise<T> {
    const result = this.tail.then(operation);
    this.tail = result.then(() => undefined, () => undefined);
    return result;
  }

  private async exchange(form: Record<string, string>): Promise<Record<string, unknown>> {
    return atStage('token_exchange',async()=>{
    const response = await this.fetcher(provider.token, {...networkOptions(), method:'POST',
      headers:{'Content-Type':'application/x-www-form-urlencoded'},
      body:new URLSearchParams({...form,resource:provider.api}).toString()});
    if (response.status === 400) {
      const error = await json(response);
      if (error.error === 'invalid_grant') throw new HttpError('sign_in_required',400);
      throw new HttpError('request_rejected',400);
    }
    return json(await checked(response));
    });
  }

  status(): Promise<{email: string; connected: boolean}> {
    return this.exclusive(async () => {
      const record = await this.store.load();
      return {email:record.session?.email ?? '',connected:!!record.session?.scope.split(' ').includes(provider.planScope)};
    });
  }

  start(): Promise<Attempt> {
    return this.exclusive(async () => begin(await this.store.load()));
  }

  finish(attempt: Attempt, url: string): Promise<void> {
    return this.exclusive(async () => {
      const {code, clientID} = callback(attempt, url);
      const record = await this.store.load();
      if (record.hostID !== attempt.registration.hostID || record.clientID !== attempt.registration.clientID) throw new Error('invalid_callback');
      const registered = {...record, clientID};
      await this.store.save(registered);
      const raw = await this.exchange({grant_type:'authorization_code',client_id:clientID,code,code_verifier:attempt.verifier,redirect_uri:attempt.redirect});
      const keys = await atStage('identity_keys',async()=>json(await checked(await this.fetcher(provider.jwks, networkOptions()))));
      const identity = await atStage('identity_verify',()=>verifyIdentity(requiredString(raw.id_token), keys, attempt, clientID));
      const session = {...tokens(raw),email:identity.email};
      requirePlan(session);
      await this.store.save({...registered,subject:identity.subject,session});
    });
  }

  access(): Promise<string> {
    return this.exclusive(async () => {
      const record = await this.store.load();
      if (!record.session || !record.clientID) throw new Error('sign_in_required');
      requirePlan(record.session);
      if (record.session.expiresAt > Date.now() + 60_000) return record.session.access;
      let raw: Record<string,unknown>;
      try {
        raw = await this.exchange({grant_type:'refresh_token',client_id:record.clientID,refresh_token:record.session.refresh});
      } catch (error) {
        if (error instanceof Error && error.message === 'sign_in_required') {
          delete record.session;
          await this.store.save(record);
        }
        throw error;
      }
      const session = {...tokens(raw,record.session),email:record.session.email};
      await this.store.save({...record,session});
      requirePlan(session);
      return session.access;
    });
  }

  logout(): Promise<void> {
    return this.exclusive(async () => {
      const record = await this.store.load();
      delete record.session;
      await this.store.save(record);
    });
  }
}
