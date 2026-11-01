export const provider = {
  issuer: 'https://auth.openai.com',
  api: 'https://api.openai.com/v1',
  token: 'https://auth.openai.com/api/accounts/oauth/token',
  jwks: 'https://auth.openai.com/.well-known/jwks.json',
  planScope: 'chatgpt.tokens.use.direct',
  redirect: 'http://127.0.0.1:54329/auth/callback',
  registrationClient: 'dynamic_agent_client',
  scope: 'openid profile email offline_access resource.invoke chatgpt.tokens.use.direct',
} as const;
export const manualInjectionMarker='shadowinput-manual-init-v1';
