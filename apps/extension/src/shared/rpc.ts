import {errorCodes} from './errors.js';
import {object} from './validation.js';
import {DiagnosedError,parseDiagnostic} from './diagnostics.js';

export async function request(type: string, fields: Record<string,unknown> = {}): Promise<unknown> {
  const reply = object(await chrome.runtime.sendMessage({type,...fields}));
  if (reply.ok !== true) {
    const code = errorCodes.find(code=>code === reply.error);
    const diagnostic=parseDiagnostic(reply.diagnostic);
    if(diagnostic && diagnostic.code===code)throw new DiagnosedError(diagnostic);
    throw new Error(code ?? 'invalid_response');
  }
  return reply.data;
}
