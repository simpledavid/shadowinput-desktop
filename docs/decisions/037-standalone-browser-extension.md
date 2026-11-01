# ADR 037: Standalone shadowinput browser extension

## Status

Accepted for preview; real OAuth/browser compatibility remains an explicit gate.

## Decision

Add `apps/extension`, a TypeScript MV3 runtime independent of Go/MyGo and the
website. Follow downward FSD imports. The worker composes account, translation
and Chrome adapters; content scripts cannot access account operations/tokens.

Use OSS PKCE/dynamic registration, verified OIDC identity and separate plan scope.
The experimental receiver handles only the created tab's top-level loopback
navigation. This differs from the documented native HTTP listener and needs
real verification. No desktop helper, copied credentials, private ChatGPT API
or API-key fallback is introduced.

Default to activeTab; request persistent permission per opted-in website.
Preserve originals and require unchanged draft snapshots. One worker owns refresh
writes, while document-scoped jobs own cancellation. Browser local storage is
not OS-keychain encryption.

## Consequences

Publish independent preview ZIPs with accurate limits. The desktop's successful
OAuth flow is not extension evidence. The full feature direction and unimplemented
slices are recorded in [the guide](../project/shadowinput-extension.md) and
[requirements](../../specs/013-browser-translation.md).
