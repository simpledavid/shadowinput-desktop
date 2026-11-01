# shadowinput

shadowinput is an AGPL-3.0 derivative of [AstraFoundry/Arion](https://github.com/AstraFoundry/Arion), based on upstream commit `b33542061b4c70c7940676f21c2a9afb6e5f1e2a` (0.1.0-alpha.20). Original copyright, license, third-party notices, source history and the maintained MyGo snapshot are retained. The source repository is [simpledavid/shadowinput-desktop](https://github.com/simpledavid/shadowinput-desktop), with `main` as its publication branch. `origin` points to that repository and `upstream` to Arion. The repository name distinguishes this desktop project from the owner's separate ShadowInput browser extension; the product name remains shadowinput.

## Downloads

The current downloadable development preview is [v0.1.0-preview.2](https://github.com/simpledavid/shadowinput-desktop/releases/tag/v0.1.0-preview.2). These are ZIP archives, not signed installers:

- [macOS Apple silicon](https://github.com/simpledavid/shadowinput-desktop/releases/download/v0.1.0-preview.2/shadowinput-v0.1.0-preview.2-macos-arm64.zip): macOS 14 or later; extract and move `shadowinput.app` to Applications. Ad hoc signed only, without Developer ID signing or Apple notarization. macOS may block opening; consult [Apple's guidance](https://support.apple.com/102445).
- [Windows x64](https://github.com/simpledavid/shadowinput-desktop/releases/download/v0.1.0-preview.2/shadowinput-v0.1.0-preview.2-windows-x64.zip): extract the entire folder and launch `shadowinput.exe`. No publisher signature; built by cross-compilation and not yet accepted on a Windows host.

The release includes SHA-256 checksums and the exact corresponding source archive. Intel Mac and Windows ARM packages are not provided. Real ChatGPT account authorization, inference and external-editor delivery remain unverified; downloading a preview does not establish those capabilities.

## Product

The [WeType comparison](../research/shadowinput-wetype.md) records competitor evidence and proposed keyboard-writing priorities separately from the implemented scope below.

- Keyboard-first desktop writing on macOS Apple silicon and Windows x64.
- Go + MyGo native UI, using Arion's Sunbeam tokens, components, icons and English/Simplified Chinese catalogs.
- Explicit polish, translation, rewrite and continuation. Continuation returns the complete draft plus its continuation; it is not inline prediction.
- Official Sign in with ChatGPT for eligible plan usage; no API key is required by this entry point.

The workflow is **select text in a supported editor → Cmd/Ctrl+Shift+J → choose task/style → Generate → review → Apply changes or Copy result**. A compact native flyout appears near the selection, with a normal window for longer editing, accounts and personal styles. Cmd/Ctrl+Enter generates; Escape cancels or closes. Unsupported selection capture shows an explicit notice; users can choose Paste original or type directly, then copy the result manually. The shortcut does not read the clipboard implicitly.

Only the visible draft, selected operation, instructions, selected style and glossary entries found in that draft are sent to OpenAI after Generate. The app starts no background clipboard collection, screen context, microphone capture, key logging, upstream accounts, telemetry, sync or updates. It does not import ChatGPT conversations or memory.

Selected text is captured through macOS Accessibility or Windows UI Automation TextPattern. The adapters retain field identity and the selected text/range, then revalidate after restoring focus and immediately before posting one paste. Changed, unavailable, protected or oversized selections are rejected. Unsupported controls remain copy-only. The target is consumed once; failures do not retry automatically. Native key-event delivery is not an atomic editor transaction and does not verify that an application consumed the paste; external undo is not implemented. The result remains on the clipboard after a paste. macOS requires Accessibility permission for selection and application; ordinary editing and copying work without it.

The result view highlights additions/deletions and can switch to plain text. A local advisory review flags changed numeric literals, recognized date expressions, protected glossary terms and some strengthened commitments. Names need explicit glossary entries; this is not general named-entity recognition. Format-preserving translation can produce false positives, and unrecognized semantic changes can be missed. Flagged results require a checkbox acknowledgment before app Copy/Apply. Large comparisons use a clearly labeled block diff. Source/task/model/instruction/style changes invalidate the previous result and review. Partial, failed and canceled generations cannot be delivered through the app's buttons.

Personal styles support name, tone, length, preferred translation language and custom rules. The glossary preserves exact names/terms, with four per-operation default style choices. Defaults are not inferred from the foreground application. Styles remain editable and are not model training. Corrupt preferences are reported rather than silently replaced.

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

The app name is `shadowinput`; the local macOS development bundle ID is `org.shadowinput.desktop.dev`. It uses the OS user-config directory's `shadowinput` folder. `host-id` is a stable opaque UUID; `preferences.json` contains only language/model choices. `writing-profile.json` stores validated style cards, glossary and task defaults as local plaintext, created with owner-only file permissions. `chatgpt.enc` is an authenticated encrypted credential snapshot, using the separate `shadowinput` / `chatgpt-vault-key` Keychain or DPAPI namespace. It does not read Arion or Codex credential files. A single-instance lock and serialized refresh prevent in-process refresh races. Sign out clears tokens while retaining the account's registration for reauthorization. Switch account clears the account-specific registration and tokens while retaining the host ID, then starts browser authorization again. The browser controls the account/workspace chosen; the app does not claim to manage browser sessions.

First sign-in dynamically registers the client using PKCE, state, nonce and an HTTP callback on `127.0.0.1` with an available port. The issued client ID is retained before exchanging the code. Eligible authorized requests use `POST https://api.openai.com/v1/responses`, `store:false`, `stream:true`; a result becomes usable only at `response.completed`. Models come from the active connection's catalog. The account card explains authorized plan use and opens the official usage page; remaining quota is not reported inside the app. Authentication/grant failures offer reauthorization, model failures offer catalog refresh, and quota failures link to usage management. Drafts are preserved and generations are never replayed automatically. Actual account eligibility, model access, limits and successful inference require a real account test.

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

Local tests exercise signed test identities, invalid state/nonce/audience/signature/expiry, declined/missing plan grants, serialized refresh rotation, truncated/error SSE responses, input limits, profile persistence/corruption, style/default deletion, diff reconstruction, advisory flags, explicit review gating and disabled delivery of partial previews. Tests use synthetic local services and do not prove real ChatGPT sign-in or plan usage. UI snapshots use synthetic content. A macOS native UI fixture confirmed scrolling, review acknowledgment enabling Copy/Apply, style-page navigation and text editing; it did not exercise model inference or cross-application selection/paste.

Required manual acceptance: complete browser login with an eligible account, verify one real rewrite, test cancellation and quota failure, grant macOS Accessibility and exercise selected-range capture, changed-range rejection and application in supported editors, then repeat tray/shortcut/selection/clipboard/credential workflows on a Windows x64 host. Headless snapshots cover both locales and light/dark modes; cross-compilation does not qualify Windows UIA behavior. Inline predictions, partial change acceptance, external undo, per-application defaults, local model integration, installers and signed updates remain follow-up work. Existing Arion features remain in the source tree and are not claimed as shadowinput features.

## License and notices

shadowinput modifications are distributed under the repository's [AGPL-3.0 license](../../LICENSE). Preserve AstraFoundry and other upstream notices. MyGo is retained with its [license](../../third_party/mygo/LICENSE); OIDC verification uses [go-oidc](https://github.com/coreos/go-oidc) and [go-jose](https://github.com/go-jose/go-jose), both Apache-2.0. Icons retain the upstream ISC/MIT notices under `ui/shared/icons/svg`. Builds stage the target's Go module dependency notices (including nested notices), Go's license/patent grant, icon notices and the root AGPL license. The shared collector is also used by the retained Arion build. No speech runtimes or bundled speech fonts are linked by this entry point. Public binary distribution must include the exact corresponding tagged shadowinput source; upstream Arion's URL alone does not provide the new modifications.

See the [writing specification](../../specs/011-writing.md) and [runtime decision](../decisions/036-shadowinput-runtime.md).
