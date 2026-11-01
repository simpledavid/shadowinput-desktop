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
- Chrome installation/UI were not verified: the browser tool explicitly blocked
  chrome://extensions. The user received the built folder and redirected work
  to website downloads instead of completing the manual installation test.
- Experimental callback reception, real account login/model listing/inference,
  browser restart/refresh/logout and live-page compatibility remain unverified.
  Desktop OAuth acceptance does not validate the extension.
- Crapper, Dryer and Mutator are not installed/integrated and were not run.
  No CRAP, duplication or mutation score is claimed.

Publish this as a preview with those limits, not a Chrome Web Store release.
