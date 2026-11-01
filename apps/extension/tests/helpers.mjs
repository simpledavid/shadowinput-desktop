export async function signingFixture() {
  const pair = await crypto.subtle.generateKey({name:'RSASSA-PKCS1-v1_5',modulusLength:2048,publicExponent:new Uint8Array([1,0,1]),hash:'SHA-256'},true,['sign','verify']);
  const keys={keys:[{...await crypto.subtle.exportKey('jwk',pair.publicKey),kid:'fixture',alg:'RS256',use:'sig'}]};
  return {keys,async token(attempt, overrides={}) {
    const encode=value=>Buffer.from(JSON.stringify(value)).toString('base64url');
    const claims={iss:'https://auth.openai.com',aud:'oaiapp_test',sub:'synthetic-user',nonce:attempt.nonce,exp:Math.floor(Date.now()/1000)+3600,email:'test@example.invalid',...overrides};
    const input=encode({alg:'RS256',kid:'fixture'})+'.'+encode(claims);
    const signature=await crypto.subtle.sign('RSASSA-PKCS1-v1_5',pair.privateKey,new TextEncoder().encode(input));
    return input+'.'+Buffer.from(signature).toString('base64url');
  }};
}
