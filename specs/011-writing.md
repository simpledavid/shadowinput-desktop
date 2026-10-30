# 011 Keyboard-first writing

This specification applies to the separate shadowinput entry point. The upstream voice entry point retains its existing behavior.

- WRT-01: Use native MyGo on macOS arm64 and Windows x64; retain AGPL and attribution.
- WRT-02: Read clipboard text only after explicit import/shortcut; submit only after Generate. Source text is bounded to 64,000 UTF-8 bytes; additional instructions to 4,000 bytes. Invalid inputs make no model request.
- WRT-03: Support polish, translate, rewrite and full-draft continuation. Preserve source independently of previews. Changed source/task/model/instructions revoke delivery of prior output.
- WRT-04: Use official OSS ChatGPT registration with fresh PKCE/state/nonce, a stable host ID and issued client ID. Verify ID-token signature, issuer, audience, expiry, nonce and subject; require the plan-use grant before inference. Keep tokens outside UI state and encrypt persisted credentials with a separate OS-protected key.
- WRT-05: Query account-specific models and stream Responses API requests. Only a nonempty `response.completed` result authorizes Copy/Paste. Errors, refusal, incomplete output, canceled jobs and late results cannot authorize delivery. Never replay an uncertain generation automatically.
- WRT-06: Paste only after an explicit click, native target checks and clipboard ownership verification. Consume the captured target and send once. Unknown delivery is never described as verified insertion. Closing cancels work and releases the target.
- WRT-07: Start no audio worker, model provisioning, upstream account, sync, usage reporting or updater. Keep normal text editing available without sign-in or Accessibility permission.
- WRT-08: Keep English/Simplified Chinese copy aligned and show actionable authentication, permission, quota, generation and clipboard errors.

Acceptance is tracked in the [shadowinput project document](../docs/project/shadowinput.md). Unit or headless tests do not establish real account or OS acceptance.
