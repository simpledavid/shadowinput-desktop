# 013: shadowinput browser translation

Download a Chrome ZIP from shadowinput.com, load unpacked, connect ChatGPT and
translate within web pages without a desktop companion. The interaction direction
follows Immersive Translate's public workflows.

| ID | Observable requirement |
| --- | --- |
| EXT-01 | ZIP contains MV3 manifest, local scripts/assets, license and installation steps. |
| EXT-02 | Fresh PKCE/state/nonce bind sign-in to the owned tab/exact callback; reject forged, expired, denied or replayed callbacks. |
| EXT-03 | Validate signed identity, issuer, audience, expiry, nonce, account and separate plan permission. |
| EXT-04 | Restore/rotate credentials; serialize refresh; never release unpersisted rotated tokens; logout clears credentials. |
| EXT-05 | Only completed responses are usable; cancellation/refusal/truncation cannot produce an applicable partial result. |
| EXT-06 | Bilingual reading preserves originals and restores them; skip hidden, editable and code content. |
| EXT-07 | Selection and Shift-hover translate requested text; render model output as inert text. |
| EXT-08 | Preview plain-input translation; reject changed drafts, passwords and unsupported rich editors. |
| EXT-09 | Site automation needs an explicit origin grant; content contexts cannot access account operations. |
| EXT-10 | Website provides separate input-method DMG/EXE and browser ZIP with accurate verification/signing limits. |

Real Chrome installation, authorization and completed inference are mandatory for
end-to-end acceptance; synthetic tests do not satisfy that gate. See
[verification](../docs/project/shadowinput-extension.md).

Subsequent slices cover TXT/subtitle files, PDF/EPUB, video adapters, site rules,
glossary/prompts, images/OCR and other platforms/engines. Preserve subtitle timing
and ebook structure; distinguish PDF extraction from layout-preserving output.
Empty controls or fake translations do not count as implemented features.
