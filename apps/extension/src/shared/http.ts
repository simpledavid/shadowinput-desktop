import { object } from './validation.js';

export type Fetcher = typeof fetch;

export async function checked(response: Response): Promise<Response> {
  if (response.ok) return response;
  if (response.status === 401) throw new Error('sign_in_required');
  if (response.status === 403) throw new Error('plan_permission');
  if (response.status === 429) throw new Error('usage_limit');
  if (response.status === 400 || response.status === 422) throw new Error('request_rejected');
  throw new Error('network_error');
}

export async function json(response: Response): Promise<Record<string, unknown>> {
  const text = await response.text();
  if (text.length > 2_000_000) throw new Error('invalid_response');
  return object(JSON.parse(text));
}

export function networkOptions(): RequestInit {
  return { credentials: 'omit', redirect: 'error', cache: 'no-store', signal: AbortSignal.timeout(20_000) };
}
