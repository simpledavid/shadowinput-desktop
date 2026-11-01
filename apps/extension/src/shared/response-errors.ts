import type {ErrorCode} from './errors.js';

const providerFailures = [
  ['subscription_sharing_usage_limit_exceeded','usage_limit'],
  ['rate_limit_exceeded','usage_limit'],
  ['subscription_sharing_usage_unavailable','service_unavailable'],
  ['subscription_sharing_user_unavailable','service_unavailable'],
  ['server_error','service_unavailable'],
  ['subscription_sharing_user_not_eligible','plan_restricted'],
  ['subscription_sharing_route_not_supported','plan_restricted'],
  ['subscription_sharing_invalid_user','plan_restricted'],
  ['chatpass_v2_scope_not_authorized','plan_restricted'],
  ['chatpass_v2_invalid_authorization_context','plan_restricted'],
  ['subscription_sharing_unsupported_capability','request_rejected'],
  ['model_not_found','model_unavailable'],
  ['model_not_available','model_unavailable'],
] as const satisfies ReadonlyArray<readonly [string,ErrorCode]>;
const streamReasons = ['stream_interrupted','response_incomplete','max_output_tokens','content_filter',
  'refusal','empty_output','output_too_large','stream_too_large','provider_failed'] as const;
export type ResponseReason = typeof providerFailures[number][0] | typeof streamReasons[number];

export class ResponseError extends Error {
  constructor(code:ErrorCode,readonly reason:ResponseReason) {super(code);}
  get kind() {return streamReasons.some(reason=>reason===this.reason)?'stream':'provider';}
}

export function responseReason(value:unknown):ResponseReason|undefined {
  return providerFailures.find(([reason])=>reason===value)?.[0] ?? streamReasons.find(reason=>reason===value);
}

export function providerFailure(value:unknown):ResponseError {
  const code=value && typeof value==='object' && 'code' in value?value.code:undefined;
  const known=providerFailures.find(([reason])=>reason===code);
  return known?new ResponseError(known[1],known[0]):new ResponseError('incomplete_response','provider_failed');
}
