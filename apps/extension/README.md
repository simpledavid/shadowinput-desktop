# shadowinput browser extension

Standalone Chrome MV3 preview. Unzip, load unpacked, then connect your own ChatGPT.
See [installation](INSTALL.txt), [verification](../../docs/project/shadowinput-extension.md),
[requirements](../../specs/013-browser-translation.md) and
[ADR 037](../../docs/decisions/037-standalone-browser-extension.md).

Toolbar actions provide bilingual main-content paragraphs, original restoration,
selection translation, optional Shift-hover and preview/apply for unchanged plain
inputs. The account page also translates typed text. Alt+Shift+T toggles page
translation; Alt+Shift+I previews the focused input. Chrome's shortcut settings
can change these bindings. Choose a model returned by the authorized account.

Reading is capped at 120 paragraphs per run with two requests per document.
Original DOM nodes remain intact. Output is inert text in isolated shadow roots.
Changed paragraphs/drafts reject stale results. Hidden content, code, password
fields and rich editors are excluded. Restore cancels that document's work.

Site automation is off by default. Enabling it requests one origin's permission.
Original text and translations are never persisted. A bounded document cache
clears on restore/navigation. No analytics, desktop helper, API-key requirement,
ChatGPT cookie access or desktop credential sharing is included.

The service worker owns credentials/refresh. Browser local/session storage is
restricted to trusted extension contexts. Local storage persists credentials
across restarts but is **not OS-keychain encryption**. Content scripts receive
only translation results and non-secret preferences.

Connection and in-page translation errors show a fixed step, failure category,
optional HTTP status and allowlisted provider/stream reason. Plan usage limits,
temporary unavailability, model restrictions, refusals, output limits and missing
completion events remain distinct. No retry or reauthorization loop is started.
Only `response.completed` can authorize a result; its final message text is used
when supplied, rather than requiring delta events. Empty/refused/incomplete final
messages are never replaced with partial streamed text.
These diagnostics exclude raw exceptions, URLs, response bodies, credentials and
translated content. The latest callback error is held only in session storage and
clears when sign-in starts/completes or the user signs out.

The [OpenAI OSS flow](https://developers.openai.com/siwc/token-sharing-open-source/sign-in)
documents an HTTP loopback listener. This experiment receives that callback via
the created authorization tab's top-level navigation event. It validates state,
PKCE, issued client ID, signed identity/nonce/account and plan scope. It never
consumes another tab's callback. This is **not documented extension support**;
real authorization and inference remain a release acceptance gate.

After replacing an unpacked extension, reload it in Chrome's extension manager
and refresh the webpage being translated so that both worker and content script
use the new build. For a failure, share only the displayed diagnostic code, not
tokens, callback URLs or a network export.

## Build

Use Node 22+ and the repository-pinned pnpm:

```sh
pnpm install --frozen-lockfile
pnpm extension:check
pnpm extension:build
```

Generated ZIP/unpacked files are in ignored `dist/`. TypeScript emits local
extension modules; esbuild bundles the isolated content script. No remote scripts
or runtime framework dependencies are loaded. SI PNG assets are transparent exports
of [the shared rounded SI vector](../shadowinput-site/public/favicon.svg), preserving
alpha at the corners. In-page branding uses the large export for high-density screens.
