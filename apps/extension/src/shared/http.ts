import { object } from './validation.js';
import {HttpError} from './errors.js';

export type Fetcher = typeof fetch;

export async function checked(response: Response): Promise<Response> {
  if (response.ok) return response;
  if (response.status === 401) throw new HttpError('sign_in_required',response.status);
  if (response.status === 403) throw new HttpError('plan_permission',response.status);
  if (response.status === 429) throw new HttpError('usage_limit',response.status);
  if (response.status === 400 || response.status === 422) throw new HttpError('request_rejected',response.status);
  throw new HttpError('network_error',response.status);
}

export async function json(response: Response): Promise<Record<string, unknown>> {
  const text = await response.text();
  if (text.length > 2_000_000) throw new Error('invalid_response');
  return object(JSON.parse(text));
}

export function networkOptions(): RequestInit {
  return { credentials: 'omit', redirect: 'error', cache: 'no-store', signal: AbortSignal.timeout(20_000) };
}
