import {parseRecord, type AccountStore} from '../features/connect/index.js';

export const ready = Promise.all([
  chrome.storage.local.setAccessLevel({accessLevel:'TRUSTED_CONTEXTS'}),
  chrome.storage.session.setAccessLevel({accessLevel:'TRUSTED_CONTEXTS'}),
]);

export const store: AccountStore = {
  async load() {
    await ready;
    const saved = await chrome.storage.local.get('account');
    if (saved.account !== undefined) return parseRecord(saved.account);
    const record = {hostID:'urn:uuid:'+crypto.randomUUID()};
    await this.save(record);
    return record;
  },
  async save(record) {
    await ready;
    await chrome.storage.local.set({account:record});
  },
};
