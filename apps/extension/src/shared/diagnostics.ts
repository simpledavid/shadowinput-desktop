import {errorCode,errorCodes,HttpError,type ErrorCode} from './errors.js';
import {ResponseError,responseReason,type ResponseReason} from './response-errors.js';

export const stages=['token_exchange','identity_keys','identity_verify','models','translation'] as const;
export type Stage=typeof stages[number];
const kinds=['http','transport','timeout','invalid_json','validation','unknown','provider','stream'] as const;
export type Diagnostic={stage:Stage;code:ErrorCode;kind:typeof kinds[number];status?:number;reason?:ResponseReason};

export class DiagnosedError extends Error {
  constructor(readonly diagnostic:Diagnostic) {super(diagnostic.code);}
}

function kind(error:unknown):Diagnostic['kind'] {
  if(error instanceof HttpError)return 'http';
  if(error instanceof ResponseError)return error.kind;
  if(!(error instanceof Error))return 'unknown';
  if(error.name==='TimeoutError')return 'timeout';
  if(error.name==='TypeError')return 'transport';
  if(error.name==='SyntaxError')return 'invalid_json';
  return errorCode(error)==='network_error'?'unknown':'validation';
}

export async function atStage<T>(stage:Stage,action:()=>Promise<T>):Promise<T> {
  try{return await action();}catch(error) {
    if(error instanceof DiagnosedError)throw error;
    const diagnostic:Diagnostic={stage,code:errorCode(error),kind:kind(error)};
    if(error instanceof HttpError)diagnostic.status=error.status;
    if(error instanceof ResponseError)diagnostic.reason=error.reason;
    // Never retain the original exception: it can contain credentials or text.
    throw new DiagnosedError(diagnostic);
  }
}

export function diagnosticOf(error:unknown):Diagnostic|undefined {
  return error instanceof DiagnosedError?error.diagnostic:undefined;
}

export function parseDiagnostic(value:unknown):Diagnostic|undefined {
  if(!value || typeof value!=='object')return;
  const stage='stage' in value?stages.find(stage=>stage===value.stage):undefined;
  const code='code' in value?errorCodes.find(code=>code===value.code):undefined;
  const kind='kind' in value?kinds.find(kind=>kind===value.kind):undefined;
  if(!stage || !code || !kind)return;
  const result:Diagnostic={stage,code,kind};
  if('status' in value && typeof value.status==='number' && Number.isInteger(value.status) && value.status>=100 && value.status<=599)result.status=value.status;
  const reason='reason' in value?responseReason(value.reason):undefined;
  if(reason)result.reason=reason;
  return result;
}

export function diagnosticText(value:unknown,labels:Record<Stage,string>):string {
  const diagnostic=parseDiagnostic(value);
  if(!diagnostic)return '';
  const status=diagnostic.status===undefined?'':` · HTTP ${diagnostic.status}`;
  const reason=diagnostic.reason?' · '+diagnostic.reason:'';
  return ` [${labels[diagnostic.stage]} · ${diagnostic.kind}${status}${reason}]`;
}
