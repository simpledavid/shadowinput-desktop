# shadowinput

shadowinput is an AGPL-3.0 derivative of [AstraFoundry/Arion](https://github.com/AstraFoundry/Arion), based on upstream commit `b33542061b4c70c7940676f21c2a9afb6e5f1e2a` (0.1.0-alpha.20). Original copyright, license, third-party notices, source history and the maintained MyGo snapshot are retained. The source repository is [simpledavid/shadowinput-desktop](https://github.com/simpledavid/shadowinput-desktop), with `main` as its publication branch. `origin` points to that repository and `upstream` to Arion. The repository name distinguishes this desktop project from the owner's separate ShadowInput browser extension; the product name remains shadowinput.

## Downloads

The first downloadable development preview is [v0.1.0-preview.1](https://github.com/simpledavid/shadowinput-desktop/releases/tag/v0.1.0-preview.1). These are ZIP archives, not signed installers:

- [macOS Apple silicon](https://github.com/simpledavid/shadowinput-desktop/releases/download/v0.1.0-preview.1/shadowinput-v0.1.0-preview.1-macos-arm64.zip): macOS 14 or later; extract and move `shadowinput.app` to Applications. Ad hoc signed only, without Developer ID signing or Apple notarization. macOS may block opening; consult [Apple's guidance](https://support.apple.com/102445).
- [Windows x64](https://github.com/simpledavid/shadowinput-desktop/releases/download/v0.1.0-preview.1/shadowinput-v0.1.0-preview.1-windows-x64.zip): extract the entire folder and launch `shadowinput.exe`. No publisher signature; built by cross-compilation and not yet accepted on a Windows host.

The release includes SHA-256 checksums and the exact corresponding source archive. Intel Mac and Windows ARM packages are not provided. Real ChatGPT account authorization, inference and external-editor delivery remain unverified; downloading a preview does not establish those capabilities.

## Product

The [WeType comparison](../research/shadowinput-wetype.md) records competitor evidence and proposed keyboard-writing priorities separately from the implemented scope below.

- Keyboard-first desktop writing on macOS Apple silicon and Windows x64.
- Go + MyGo native UI, using Arion's Sunbeam tokens, components, icons and English/Simplified Chinese catalogs.
- Explicit polish, translation, rewrite and continuation. Continuation returns the complete draft plus its continuation; it is not inline prediction.
- Official Sign in with ChatGPT for eligible plan usage; no API key is required by this entry point.

The first workflow is **copy text → Cmd/Ctrl+Shift+J → choose a task → Generate → review → Copy result or Paste back**. Users can also type or paste directly into the window. Only the visible draft, chosen operation and additional instructions are sent to OpenAI after Generate. Clipboard collection, screen context, microphone capture, key logging, upstream accounts, telemetry, sync and upstream updates are not started. A normal window, menu-bar/system-tray entry and global shortcut provide access.

Paste back captures the external target before the editor opens, reactivates it and reuses Arion's native focus/security/clipboard-revision checks. It posts one paste at the target's current selection and consumes that target. It does not guarantee an unchanged original text range or verify that an application consumed the paste. Failures do not retry automatically. The result remains on the clipboard after a paste. macOS requires Accessibility permission for this action; ordinary editing and copying work without it. Source edits or operation changes invalidate the result. Partial, failed and canceled generations cannot be delivered through the app's Copy/Paste buttons.

## Reuse and new modules

| Reused upstream component | Purpose |
| --- | --- |
| `third_party/mygo` | Native window, tray, shortcut, clipboard and editing |
| `ui/shared/kit`, `theme`, `icons`, `copy` | UI and localization |
| `infrastructure/clipboardos`, `macos`, `windows` | Guarded external-target delivery |
| `internal/oauthloopback`, `infrastructure/browser` | PKCE, callback listener and system browser |
| `infrastructure/encryptedfile`, `secrets` | AES-GCM snapshot and Keychain/DPAPI-protected key |

The new `cmd/shadowinput` entry point composes `internal/shadowinput`, `ui/pages/writing`, `ui/features/writing` and `ui/shared/writing`. Portable writing validation/instructions and the generator port live in core. `infrastructure/chatgptplan` implements the documented OSS flow and public Responses endpoint, independently of Arion's existing Codex-backend provider. `coreos/go-oidc/v3` verifies the ID token's signature, issuer, audience and expiry; the adapter also verifies nonce, subject and plan grant.

## Identity and storage

The app name is `shadowinput`; the local macOS development bundle ID is `org.shadowinput.desktop.dev`. It uses the OS user-config directory's `shadowinput` folder. `host-id` is a stable opaque UUID; `preferences.json` contains only language/model choices. `chatgpt.enc` is an authenticated encrypted credential snapshot, using the separate `shadowinput` / `chatgpt-vault-key` Keychain or DPAPI namespace. It does not read Arion or Codex credential files. A single-instance lock and serialized refresh prevent in-process refresh races. Sign out clears tokens while retaining the account's registration for reauthorization. Switching to a different ChatGPT account/workspace is not yet exposed.

First sign-in dynamically registers the client using PKCE, state, nonce and an HTTP callback on `127.0.0.1` with an available port. The issued client ID is retained before exchanging the code. Eligible authorized requests use `POST https://api.openai.com/v1/responses`, `store:false`, `stream:true`; a result becomes usable only at `response.completed`. Models come from the active connection's catalog. Actual account eligibility, model access, limits and successful inference require a real account test.

References: [official registration](https://developers.openai.com/siwc/token-sharing-open-source/sign-in), [models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference), [sessions](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions), [Paste Perfect cookbook](https://developers.openai.com/cookbook/articles/sign-in-with-chatgpt).

## Development

Use the repository's Go toolchain and pnpm version. The writing app does not stage or download speech/model runtimes.

```sh
node scripts/shadowinput.mjs check
node scripts/shadowinput.mjs dev
node scripts/shadowinput.mjs build macos
node scripts/shadowinput.mjs build windows
```

macOS builds require Xcode command-line tools and produce an ad hoc signed `.app` under `apps/desktop/build/shadowinput/macos`. Windows builds produce a standalone x64 `.exe` under the corresponding `windows` directory; it can be cross-built from macOS. Ad hoc signing and cross-compilation do not establish distribution readiness or Windows native acceptance. No installer, Developer ID notarization, Windows publisher signing or update channel has been configured for shadowinput.

GitHub Actions is disabled in this source repository while the retained Arion workflows await adaptation to the independent shadowinput entry point. Local checks and builds remain available through the commands above.

For a release, run checks, commit the source, and create the intended version tag at `HEAD`. Then use `node scripts/shadowinput.mjs package macos <tag>` and `node scripts/shadowinput.mjs package windows <tag>`. Packaging requires a clean checkout and matching tag, rebuilds each target into a fresh directory, records the source commit and actual Go toolchain in `BUILD-INFO.json`, and produces ZIP archives plus individual checksums under `apps/desktop/build/shadowinput/releases`. Packaging is supported on macOS and Windows; macOS builds require a Mac. Publish the tagged corresponding source (`git archive`), combined checksums, and both ZIPs together as a GitHub prerelease. No automatic updater is enabled.

## Validation and remaining acceptance

Local tests exercise signed test identities, invalid state/nonce/audience/signature/expiry, declined/missing plan grants, serialized refresh rotation, truncated/error SSE responses, input limits and disabled delivery of partial previews. Tests use synthetic local services and do not prove real ChatGPT sign-in or plan usage. UI snapshots use synthetic content.

Required manual acceptance: complete browser login with an eligible account, verify one real rewrite, test cancellation and quota failure, grant macOS Accessibility and paste into supported editors, then repeat tray/shortcut/clipboard/credential workflows on a Windows x64 host. Broader direct-selection capture, inline predictions, undo of external replacement, personal vocabulary integration, account switching, installers and signed updates remain follow-up work. Existing Arion features remain in the source tree and are not claimed as shadowinput features.

## License and notices

shadowinput modifications are distributed under the repository's [AGPL-3.0 license](../../LICENSE). Preserve AstraFoundry and other upstream notices. MyGo is retained with its [license](../../third_party/mygo/LICENSE); OIDC verification uses [go-oidc](https://github.com/coreos/go-oidc) and [go-jose](https://github.com/go-jose/go-jose), both Apache-2.0. Icons retain the upstream ISC/MIT notices under `ui/shared/icons/svg`. Builds stage the target's Go module dependency notices (including nested notices), Go's license/patent grant, icon notices and the root AGPL license. The shared collector is also used by the retained Arion build. No speech runtimes or bundled speech fonts are linked by this entry point. Public binary distribution must include the exact corresponding tagged shadowinput source; upstream Arion's URL alone does not provide the new modifications.

See the [writing specification](../../specs/011-writing.md) and [runtime decision](../decisions/036-shadowinput-runtime.md).
