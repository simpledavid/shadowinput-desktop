# shadowinput

shadowinput is an AGPL-3.0 derivative of [AstraFoundry/Arion](https://github.com/AstraFoundry/Arion), based on upstream commit `b33542061b4c70c7940676f21c2a9afb6e5f1e2a` (0.1.0-alpha.20). Original copyright, license, third-party notices, source history and the maintained MyGo snapshot are retained. The source repository is [simpledavid/shadowinput-desktop](https://github.com/simpledavid/shadowinput-desktop), with `main` as its publication branch. `origin` points to that repository and `upstream` to Arion. The repository name distinguishes this desktop project from the owner's separate ShadowInput browser extension; the product name remains shadowinput.

Development and review follow the [ub-mode coding rules](../quality/ub-mode.md).

## Downloads

The current downloadable development preview is [v0.1.0-preview.5](https://github.com/simpledavid/shadowinput-desktop/releases/tag/v0.1.0-preview.5). Use the DMG or EXE installer; portable ZIP archives remain available on the release page:

- [macOS Apple silicon](https://github.com/simpledavid/shadowinput-desktop/releases/download/v0.1.0-preview.5/shadowinput-v0.1.0-preview.5-macos-arm64.dmg): macOS 14 or later; open the disk image and drag `shadowinput.app` to Applications. Ad hoc signed only, without Developer ID signing or Apple notarization. macOS may block opening; consult [Apple's guidance](https://support.apple.com/102445).
- [Windows x64](https://github.com/simpledavid/shadowinput-desktop/releases/download/v0.1.0-preview.5/shadowinput-v0.1.0-preview.5-windows-x64-setup.exe): run the installer, then open shadowinput from the Start menu. Installs for the current user without elevation; all local-runtime files are included. Uninstall through Windows Settings preserves downloaded models, preferences and credentials. No publisher signature; built by cross-compilation and not yet accepted on a Windows host.

The release includes SHA-256 checksums and the exact corresponding source archive. Intel Mac and Windows ARM packages are not provided. ChatGPT authorization, session restoration and one completed real rewrite have been verified on macOS with one account. Native Windows behavior and external-editor delivery remain unverified; downloading a preview does not establish those capabilities.

## Product

The current development branch also contains an experimental **system Pinyin input method with optional AI candidates**, with setup as the default app page. See [input-method workflow, build prerequisites and acceptance status](shadowinput-ime.md). Installed macOS source discovery, enablement and selection are confirmed; full editor/AI candidate acceptance remains pending. Windows TSF is cross-built but not native-tested. These changes are not in the preview.5 downloads above. The following writing workflow remains available as the separate long-form tool. Its Generate-only privacy description applies to that tool; enabling automatic cloud candidates also sends the current composition as described in the input-method guide.

The [WeType comparison](../research/shadowinput-wetype.md) records competitor evidence and proposed keyboard-writing priorities separately from the implemented scope below.

- Keyboard-first desktop writing on macOS Apple silicon and Windows x64.
- Go + MyGo native UI, using Arion's Sunbeam tokens, components, icons and English/Simplified Chinese catalogs.
- Explicit polish, translation, rewrite and continuation. Continuation returns the complete draft plus its continuation; it is not inline prediction.
- Offline writing with Qwen3.5-2B Q4_K_M, without login or an API key.
- Official Sign in with ChatGPT for eligible plan usage; no API key is required by this entry point.

The workflow is **select text in a supported editor → Cmd/Ctrl+Shift+J → choose task/style → Generate → review → Apply changes or Copy result**. A compact native flyout appears near the selection, with a normal window for longer editing, accounts and personal styles. Cmd/Ctrl+Enter generates; Escape cancels or closes. Unsupported selection capture shows an explicit notice; users can choose Paste original or type directly, then copy the result manually. The shortcut does not read the clipboard implicitly.

In ChatGPT mode, only the visible draft, selected operation, instructions, selected style and glossary entries found in that draft are sent to OpenAI after Generate. In local mode they remain on this computer; switching to cloud always requires an explicit provider choice. The app starts no background clipboard collection, screen context, microphone capture, key logging, upstream accounts, telemetry, sync or updates. It does not import ChatGPT conversations or memory.

Selected text is captured through macOS Accessibility or Windows UI Automation TextPattern. The adapters retain field identity and the selected text/range, then revalidate after restoring focus and immediately before posting one paste. Changed, unavailable, protected or oversized selections are rejected. Unsupported controls remain copy-only. The target is consumed once; failures do not retry automatically. Native key-event delivery is not an atomic editor transaction and does not verify that an application consumed the paste; external undo is not implemented. The result remains on the clipboard after a paste. macOS requires Accessibility permission for selection and application; ordinary editing and copying work without it.

The result view highlights additions/deletions and can switch to plain text. A local advisory review flags changed numeric literals, recognized date expressions, protected glossary terms and some strengthened commitments. Names need explicit glossary entries; this is not general named-entity recognition. Format-preserving translation can produce false positives, and unrecognized semantic changes can be missed. Flagged results require a checkbox acknowledgment before app Copy/Apply. Large comparisons use a clearly labeled block diff. Source/task/provider/model/instruction/style changes invalidate the previous result and review. Partial, failed and canceled generations cannot be delivered through the app's buttons.

Personal styles support name, tone, length, preferred translation language and custom rules. The glossary preserves exact names/terms, with four per-operation default style choices. Defaults are not inferred from the foreground application. Styles remain editable and are not model training. Corrupt preferences are reported rather than silently replaced.

## Local model

Choose **Local model → Download & start**. The first download is 1,280,835,840 bytes (about 1.28 GB); existing resumable data is reused. Downloaded weights stay under the app’s `models` directory and are checked against the pinned size and SHA-256 before use. Later **Start model** works offline; **Stop model** releases the worker and its native child without deleting weights. Cancel setup stops the current request and does not resume it automatically. Selecting local or reopening the app alone never starts a download. New installations default to local; existing saved ChatGPT preferences retain their provider.

The selected artifact is [Qwen3.5-2B Q4_K_M](https://huggingface.co/unsloth/Qwen3.5-2B-GGUF/tree/f6d5376be1edb4d416d56da11e5397a961aca8ae), a community quantization of [Qwen3.5-2B](https://huggingface.co/Qwen/Qwen3.5-2B), under Apache-2.0. Its legacy catalog ID ends in `-metal` but also identifies the Windows CPU artifact. 4B, 9B, arbitrary GGUF imports and Ollama are not integrated. Model downloads may use the existing Hugging Face mirror/retry transport; writing text is never part of that download traffic.

Packages include the pinned llama.cpp b11734 runtime: Metal on Apple silicon, CPU baseline/AVX2 on Windows x64. The worker selects the Windows executable after CPU/OS feature checks. It uses a private Unix socket, one slot and four decode threads, without a public TCP listener or disk prompt cache. Thinking is disabled. The full native chat template and requested output must fit 4096 tokens; output is bounded to 512–2048 tokens depending on source length, with a 90-second app deadline. Oversized requests and incomplete responses preserve the original draft and give an explicit error. Cancellation clears the current response and destroys native state; the next explicit generation can reload it. Styles and review work identically across providers.

This small model may invent facts, especially while continuing a draft. A smoke test changed a request for confirmation into a statement that it was confirmed; advisory checks do not guarantee detection of every semantic change. Always review the complete preview. Speed and memory depend on the computer; no minimum-memory or quality benchmark is claimed.

## Reuse and new modules

| Reused upstream component | Purpose |
| --- | --- |
| `third_party/mygo` | Native window, tray, shortcut, clipboard and editing |
| `ui/shared/kit`, `theme`, `icons`, `copy` | UI and localization |
| `infrastructure/clipboardos`, `macos`, `windows` | Guarded external-target delivery |
| `infrastructure/provision`, `rewriting`, `llm` | Pinned model download, worker lifecycle and offline completion |
| `internal/oauthloopback`, `infrastructure/browser` | PKCE, callback listener and system browser |
| `infrastructure/encryptedfile`, `secrets` | AES-GCM snapshot and Keychain/DPAPI-protected key |

The new `cmd/shadowinput` entry point composes `internal/shadowinput`, `ui/pages/writing`, `ui/features/writing` and `ui/shared/writing`. Portable writing validation/instructions and the generator port live in core. `infrastructure/chatgptplan` implements the documented OSS flow and public Responses endpoint, independently of Arion's existing Codex-backend provider. `coreos/go-oidc/v3` verifies the ID token's signature, issuer, audience and expiry; the adapter also verifies nonce, subject and plan grant.

## Identity and storage

The yellow SI mark is shared by the desktop header, Dock/window/taskbar, tray, input-method host, Windows installer/uninstaller and website. Its canonical vector is [mark.svg](../../apps/desktop/ui/shared/shadowbrand/mark.svg). Run `go run ./apps/desktop/cmd/shadowinput-icon` from the repository root to export PNG, ICO, ICNS and web assets; `-check` verifies that they match. The upstream Arion entry point and dependency logos retain their own branding. Existing published release assets are immutable; the unified icons ship in subsequent builds.

The app name is `shadowinput`; the local macOS development bundle ID is `org.shadowinput.desktop.dev`. It uses the OS user-config directory's `shadowinput` folder. `host-id` is a stable opaque UUID; `preferences.json` contains language/provider/model choices. Local mode does not initialize the ChatGPT credential vault. `writing-profile.json` stores validated style cards, glossary and task defaults as local plaintext, created with owner-only file permissions. `chatgpt.enc` is an authenticated encrypted credential snapshot, using the separate `shadowinput` / `chatgpt-vault-key` Keychain or DPAPI namespace. It does not read Arion or Codex credential files. A single-instance lock and serialized refresh prevent in-process refresh races. Sign out clears tokens while retaining the account's registration for reauthorization. Switch account clears the account-specific registration and tokens while retaining the host ID, then starts browser authorization again. The browser controls the account/workspace chosen; the app does not claim to manage browser sessions.

First sign-in dynamically registers the client using PKCE, state, nonce and an HTTP callback on `127.0.0.1` with an available port. The issued client ID is retained before exchanging the code. Eligible authorized requests use `POST https://api.openai.com/v1/responses`, `store:false`, `stream:true`, and an `input` message array (the plan route rejects the regular API string shorthand); a result becomes usable only at `response.completed`. Models come from the active connection's catalog. The account card explains authorized plan use and opens the official usage page; remaining quota is not reported inside the app. Authentication/grant failures offer reauthorization, model failures offer catalog refresh, and quota failures link to usage management. Drafts are preserved and generations are never replayed automatically. Eligibility, model access and limits depend on the selected account. On 2026-11-01, a real macOS account completed authorization, model listing, session restoration after restart and one rewrite through `response.completed`. This does not establish eligibility for every account or native Windows behavior.

References: [plan request limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations), [official registration](https://developers.openai.com/siwc/token-sharing-open-source/sign-in), [models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference), [sessions](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions), [Paste Perfect cookbook](https://developers.openai.com/cookbook/articles/sign-in-with-chatgpt).

## Development

Use the repository's Go toolchain and pnpm version. Development and packaging stage the pinned local writing runtime and rebuild its Go worker; no speech runtime or model weights are downloaded by these commands.

```sh
node scripts/shadowinput.mjs check
node scripts/shadowinput.mjs dev
node scripts/shadowinput.mjs build macos
node scripts/shadowinput.mjs build windows
```

macOS builds require Xcode command-line tools and produce an ad hoc signed `.app` under `apps/desktop/build/shadowinput/macos`. Windows builds produce an x64 executable with an adjacent `llm` runtime folder under the corresponding `windows` directory; keep the entire folder together. Windows runtime builds require CMake and a C++ toolchain (MSVC on Windows, MinGW-w64 for cross-compilation from macOS). Ad hoc signing and cross-compilation do not establish distribution readiness or Windows native acceptance. Packaging adds a macOS DMG and a per-user Windows EXE installer (NSIS 3, with a Start menu shortcut and registered uninstaller). The installer requires Windows 10 or newer on x64, checks native payload files are not in use, and removes only packaged files when uninstalled. It preserves user configuration and downloaded models. Developer ID notarization, Windows publisher signing and an update channel are not configured.

GitHub Actions is disabled in this source repository while the retained Arion workflows await adaptation to the independent shadowinput entry point. Local checks and builds remain available through the commands above.

For a release, run checks, commit the source, and create the intended version tag at `HEAD`. Then use `node scripts/shadowinput.mjs package macos <tag>` and `node scripts/shadowinput.mjs package windows <tag>`. Packaging requires a clean checkout and matching tag, rebuilds each target into its own `apps/desktop/build/shadowinput/<tag>/<target>` directory, records the source commit, pinned runtime/model identity and actual Go toolchain in `BUILD-INFO.json`, and produces DMG/EXE installers, portable ZIP archives and individual checksums under `apps/desktop/build/shadowinput/releases`. Packaging is supported on macOS and Windows; macOS builds require a Mac. Windows packaging also requires NSIS 3 (`makensis` on PATH, or the `MAKENSIS` environment variable). On macOS it can be installed with Homebrew: `brew install makensis`. Installer inventory checks run with `node --test scripts/lib/shadowinput-installers.test.mjs`. Publish the tagged corresponding source (`git archive`), combined checksums, both installers and both ZIPs together as a GitHub prerelease. A running macOS build is never overwritten; quit it before rebuilding the same path. No automatic updater is enabled.

## Validation and remaining acceptance

Local tests exercise signed test identities, invalid state/nonce/audience/signature/expiry, declined/missing plan grants, serialized refresh rotation, truncated/error SSE responses, input limits, profile persistence/corruption, style/default deletion, diff reconstruction, advisory flags, explicit review gating and disabled delivery of partial previews. These unit tests use synthetic local services. Separately, the 2026-11-01 macOS smoke test above verified one real account and completed plan inference. The plan request regression requires an `input` array and accepts an absent SSE Content-Type header only when the response body passes the completion gate. UI snapshots use synthetic content. A macOS native UI fixture confirmed scrolling, review acknowledgment enabling Copy/Apply, style-page navigation and text editing; it did not exercise model inference or cross-application selection/paste.

Local writing tests also cover provider isolation, legacy preferences, explicit download/cancel states, bounded output, typed worker errors and slot erasure. Real macOS Metal inference passed all four writing operations, context-overflow rejection, cancellation and a subsequent fresh request. These authored smoke cases are integration evidence, not a quality benchmark. Windows CPU executables were cross-built and their DLL closure checked; they still need a native host test.

Remaining manual acceptance: test real cancellation and quota failure, grant macOS Accessibility and exercise selected-range capture, changed-range rejection and application in supported editors, then repeat installer/upgrade/uninstall, tray/shortcut/selection/clipboard/credential workflows on a Windows x64 host. Headless snapshots cover both locales and light/dark modes; cross-compilation does not qualify Windows UIA behavior. Inline predictions, partial change acceptance, external undo, per-application defaults, larger/custom local models and signed updates remain follow-up work. Existing Arion features remain in the source tree and are not claimed as shadowinput features.

## License and notices

shadowinput modifications are distributed under the repository's [AGPL-3.0 license](../../LICENSE). Preserve AstraFoundry and other upstream notices. MyGo is retained with its [license](../../third_party/mygo/LICENSE); OIDC verification uses [go-oidc](https://github.com/coreos/go-oidc) and [go-jose](https://github.com/go-jose/go-jose), both Apache-2.0. Icons retain the upstream ISC/MIT notices under `ui/shared/icons/svg`. Builds stage the target's Go module dependency notices (including nested notices), Go's license/patent grant, icon notices and the root AGPL license. The shared collector is also used by the retained Arion build. Packages include llama.cpp and dependency notices; Windows MinGW cross-builds also retain GCC runtime-exception and MinGW/winpthreads notices from [runtime notices](../../third_party/runtime-notices/README.md). No speech runtimes or bundled speech fonts are linked by this entry point. Public binary distribution must include the exact corresponding tagged shadowinput source; upstream Arion's URL alone does not provide the new modifications.

See the [writing specification](../../specs/011-writing.md) and [runtime decision](../decisions/036-shadowinput-runtime.md).
