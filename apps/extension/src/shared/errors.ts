export const errorCodes = [
  'sign_in_required','plan_permission','usage_limit','request_rejected','network_error','invalid_response',
  'storage_failed','invalid_callback','authorization_denied','auth_expired','invalid_identity','account_mismatch',
  'incomplete_response','invalid_input','cancelled','busy','unsupported_page','site_permission','draft_changed',
  'service_unavailable','plan_restricted','model_unavailable','translation_refused','output_limit',
] as const;
export type ErrorCode = typeof errorCodes[number];
export class HttpError extends Error {
  constructor(code: ErrorCode, readonly status: number) { super(code); }
}
export function errorCode(error: unknown): ErrorCode {
  if (error instanceof Error) {
    for (const code of errorCodes) if (error.message === code) return code;
    if (error.name === 'AbortError') return 'cancelled';
    if (error.name === 'SyntaxError') return 'invalid_response';
  }
  return 'network_error';
}
