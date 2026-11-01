# shadowinput browser translation

The user requested a standalone Chrome ZIP, their own ChatGPT account, and both
products on shadowinput.com. Interaction direction follows the public
[Immersive Translate feature documentation](https://immersivetranslate.com/zh-Hans/docs/).
Source code, interface and branding are shadowinput's own implementation.

## Implemented preview

- Authorization, verified identity/plan grant, rotating refresh, local logout,
  authorized model discovery and explicit text translation.
- Bilingual paragraphs, restore, selection, optional Shift-hover and preview/apply
  for unchanged plain inputs.
- Optional automatic translation per explicitly authorized website.
- Chinese/English UI; Chinese/English/Japanese targets; cancellation and recovery.

This is not production feature parity. PDF layout translation, EPUB export,
subtitle files/live video, image/OCR, glossary/prompts, more engines, mobile
browsers and site-specific video adapters are not implemented in this ZIP.
They remain subsequent slices after real account/browser acceptance.

The [extension README](../../apps/extension/README.md) owns installation,
permissions and storage details. Text is sent after an explicit action, or
automatically for an enabled site. No shadowinput proxy, analytics, chat-history
access or persistent original/translation history is used. Local logout clears
credentials; it does not revoke the provider-side grant.

## Evidence and limits

- Strict TypeScript compilation, synthetic signed-token/transport tests, DOM
  reading/replacement tests, worker sender/callback ownership and ZIP build ran.
  Test output owns the current count; fixtures are not real inference evidence.
- In the initial attempt, the user loaded the extension and reported an error after
  completing OpenAI authorization and returning to the account page. This is a
  failed real acceptance attempt, not successful sign-in. The browser tool blocks
  both chrome://extensions and extension pages, so their UI requires user checks.
- Safe diagnostics identify token exchange, identity keys/verification, model
  discovery and translation separately. Only allowlisted stage/code/kind and an
  optional HTTP status cross the worker boundary. Callback failure metadata lives
  in session storage and clears with a new sign-in, successful sign-in or logout.
  Raw errors, URLs, response bodies, credentials and text are excluded.
- The first callback failure was narrowed to token-exchange transport. A Chrome
  Worker probe reproduced `Illegal invocation` when native fetch was called as an
  account object's method; unbound/global-bound calls succeeded with the same
  fixed data URL. The default account transport now binds fetch to the global
  receiver, covered by a failing-then-passing regression. This browser-runtime
  reproduction does not itself establish successful OpenAI inference.
- After the fetch-receiver fix, a user-provided screenshot showed the connected
  account and an authorized model in both the workbench and popup. Real sign-in
  and model discovery now have manual evidence; completed translation is pending.
- Completed real translation, browser restart/token rotation/logout and live-page
  compatibility remain unverified.
  Desktop OAuth acceptance does not validate the extension.
- Pre-merge review reproduced and fixed three regressions: an old model request
  could clear a newer login, the popup enabled page actions on unsupported pages,
  and closing the text workbench left its translation running. Worker and DOM
  regression tests now check session ownership, disabled actions with guidance,
  and cancellation when the owning tab closes. These are synthetic checks.
- Crapper, Dryer and Mutator are not installed/integrated and were not run.
  No CRAP, duplication or mutation score is claimed.

Publish this as a preview with those limits, not a Chrome Web Store release.
