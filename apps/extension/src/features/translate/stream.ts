import { object } from '../../shared/validation.js';

function eventData(frame: string): Record<string, unknown> | undefined {
  const data = frame.split('\n').filter(line => line.startsWith('data:')).map(line => line.slice(5).trimStart()).join('\n');
  return data ? object(JSON.parse(data)) : undefined;
}

function consume(event: Record<string, unknown>, text: string): {text: string; completed: boolean} {
  if (event.type === 'response.output_text.delta') {
    if (typeof event.delta !== 'string' || text.length + event.delta.length > 256_000) throw new Error('incomplete_response');
    return {text: text + event.delta, completed:false};
  }
  if (event.type === 'response.completed') {
    if (object(event.response).status !== 'completed' || !text.trim()) throw new Error('incomplete_response');
    return {text,completed:true};
  }
  if (['response.incomplete','response.failed','error','response.refusal.delta','response.refusal.done'].includes(String(event.type))) throw new Error('incomplete_response');
  return {text,completed:false};
}

export async function readTranslation(response: Response): Promise<string> {
  if (!response.body) throw new Error('incomplete_response');
  const reader = response.body.getReader();
  const decoder = new TextDecoder('utf-8',{fatal:true});
  let buffer = '', text = '', total = 0;
  try {
    while (true) {
      const chunk = await reader.read();
      if (chunk.done) throw new Error('incomplete_response');
      total += chunk.value.length;
      if (total > 4_000_000) throw new Error('incomplete_response');
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
