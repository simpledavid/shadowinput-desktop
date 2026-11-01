# ADR 036: Independent keyboard-first shadowinput runtime

Status: Accepted for this derivative.

## Decision

Keep Arion's source and native MyGo framework. Add a separate composition root for shadowinput instead of starting the voice bootstrap. Reuse native clipboard-target adapters, UI primitives, browser/loopback helpers, encryption and OS key storage. Use the official Sign in with ChatGPT OSS dynamic-registration flow and public Responses API for the new writing provider. Verify OIDC with the maintained go-oidc library. The existing Arion cloud provider remains unchanged.

Use a separate app name, configuration path, single-instance identity, encrypted credential namespace and macOS development bundle ID. Do not inherit the upstream update, account or telemetry startup. The shared workflow captures a supported selected range or imports explicitly copied text. The separate native input-method entry point owns automatic phrase suggestions; it does not use clipboard capture.

## Consequences

The derivative omits the ASR runtime. Local writing reuses Arion’s provision store, pinned GGUF worker and private-socket llama.cpp runtime; packages include that runtime but download weights only on explicit request. Both providers share core writing instructions, styles and review. Provider selection never silently changes after an error. Clipboard paste is a one-shot user request, with original upstream validation limits. macOS and Windows compile targets share the UI and auth code, but both require native acceptance. Browser login is a user-controlled step. See [shadowinput](../project/shadowinput.md) and [011](../../specs/011-writing.md).

## Native phrase continuation (2026-11-01)

Normal Rime word commits stay synchronous and independent of the model. After a short pause, opt-in suggestions use the resident local worker only, even when ChatGPT is selected for the writing tool. Retain at most 128 characters of this input method’s contiguous commits for 30 seconds, tied to the active client and caret. Tab appends at the unchanged caret; typing, deletion, navigation, focus changes or opt-out invalidate pending work. Suggestions never replace committed text or enter the learning dictionary.

Download remains explicit. An opted-in user with a verified installed model gets a warm worker after app startup. Missing, slow or invalid results leave normal input available; there is no automatic cloud fallback. Local completion has a distinct suffix-only task, a short output budget, and a two-second deadline. The writing tool retains its independent provider selection and full-result review.
