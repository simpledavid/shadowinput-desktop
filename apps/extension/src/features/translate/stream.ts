import { object } from '../../shared/validation.js';
import {ResponseError,providerFailure} from '../../shared/response-errors.js';

function bounded(text:string):string {
  if(text.length>256_000)throw new ResponseError('incomplete_response','output_too_large');
  return text;
}

function textPart(raw:unknown):string {
  const part=object(raw);
  if(part.type==='refusal')throw new ResponseError('translation_refused','refusal');
  if(part.type!=='output_text')return '';
  if(typeof part.text!=='string')throw new Error('invalid_response');
  return bounded(part.text);
}

function completedMessage(raw:unknown):string {
  const message=object(raw);
  if(message.type!=='message' || message.role!=='assistant')return '';
  if(message.status!==undefined && message.status!=='completed')throw new ResponseError('incomplete_response','response_incomplete');
  if(!Array.isArray(message.content))throw new Error('invalid_response');
  const content=bounded(message.content.map(textPart).join(''));
  return message.phase==='commentary'?'':content;
}

function completedText(response:Record<string,unknown>,delta:string):string {
  if(response.status!=='completed')throw new ResponseError('incomplete_response','response_incomplete');
  if(response.error)throw providerFailure(response.error);
  let text=delta;
  if(response.output!==undefined) {
    if(!Array.isArray(response.output))throw new Error('invalid_response');
    text=response.output.map(completedMessage).filter(Boolean).join('\n');
  }
  if(!text.trim())throw new ResponseError('incomplete_response','empty_output');
  return bounded(text);
}

function eventData(frame: string): Record<string, unknown> | undefined {
  const data = frame.split('\n').filter(line => line.startsWith('data:')).map(line => line.slice(5).trimStart()).join('\n');
  if(data==='[DONE]')throw new ResponseError('incomplete_response','stream_interrupted');
  return data ? object(JSON.parse(data)) : undefined;
}

function consume(event: Record<string, unknown>, text: string): {text: string; completed: boolean} {
  if (event.type === 'response.output_text.delta') {
    if (typeof event.delta !== 'string') throw new Error('invalid_response');
    return {text: bounded(text + event.delta), completed:false};
  }
  if (event.type === 'response.completed') {
    return {text:completedText(object(event.response),text),completed:true};
  }
  if(event.type==='response.failed')throw providerFailure(event.response?object(event.response).error:undefined);
  if(event.type==='error')throw providerFailure(event.error ?? event);
  if(event.type==='response.incomplete') {
    const details=event.response?object(event.response).incomplete_details:undefined;
    const reason=details?object(details).reason:undefined;
    if(reason==='max_output_tokens')throw new ResponseError('output_limit',reason);
    if(reason==='content_filter')throw new ResponseError('translation_refused',reason);
    throw new ResponseError('incomplete_response','response_incomplete');
  }
  if (['response.refusal.delta','response.refusal.done'].includes(String(event.type)))throw new ResponseError('translation_refused','refusal');
  return {text,completed:false};
}

export async function readTranslation(response: Response): Promise<string> {
  if (!response.body) throw new ResponseError('incomplete_response','stream_interrupted');
  const reader = response.body.getReader();
  const decoder = new TextDecoder('utf-8',{fatal:true});
  let buffer = '', text = '', total = 0;
  try {
    while (true) {
      const chunk = await reader.read();
      if (chunk.done) throw new ResponseError('incomplete_response','stream_interrupted');
      total += chunk.value.length;
      if (total > 4_000_000) throw new ResponseError('incomplete_response','stream_too_large');
      buffer += decoder.decode(chunk.value,{stream:true});
      buffer = buffer.replaceAll('\r\n','\n');
      let end: number;
      while ((end = buffer.indexOf('\n\n')) >= 0) {
        const event = eventData(buffer.slice(0,end));
        buffer = buffer.slice(end+2);
        if (!event) continue;
        const result = consume(event,text);
        text = result.text;
        if (result.completed) return text;
      }
    }
  } finally {
    await reader.cancel();
    reader.releaseLock();
  }
}
