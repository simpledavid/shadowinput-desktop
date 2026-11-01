# shadowinput and WeType

This study applies to the [shadowinput derivative](../project/shadowinput.md), built from Arion with Go and MyGo for macOS and Windows. It records vendor-documented capabilities, source-confirmed shadowinput behavior and proposed product priorities separately. Proposed features below are not implemented acceptance claims.

## Positioning recommendation

Build a keyboard-first writing companion for people who already use ChatGPT and frequently write messages, email and bilingual work documents. The proposed promise is: **Use your own ChatGPT account to write in your own voice, wherever you type.**

ChatGPT sign-in reduces setup friction for eligible users. Lasting value should come from fewer interactions, controlled edits and reusable personal preferences. Model availability alone does not establish better writing quality. Measure the complete task from draft to accepted text.

The initial companion can coexist with an existing input method, including WeType. Native IME integration remains a later product step using Arion's existing platform work, subject to separate macOS and Windows acceptance. The current shadowinput entry point does not activate that IME.

## Vendor evidence

Evidence comes from WeType's public website, its embedded `window.injectData.appChangelog` records and the corresponding public change-log routes. The browser extraction service could not open these pages; direct HTTPS requests returned the official pages successfully. This is documentation research, not a hands-on competitor performance test. Historical release notes establish announced capability, not its unchanged behavior in every current build.

| Area | Documented WeType capability | Product implication |
| --- | --- | --- |
| Keyboard foundations | Pinyin configuration, double Pinyin/Wubi, phrase tools, punctuation and candidate controls appear in desktop release notes. [Mac overview](https://z.weixin.qq.com/web/change-log/8), [double Pinyin/Wubi](https://z.weixin.qq.com/web/change-log/26), [Windows](https://z.weixin.qq.com/web/change-log/29) | Typing stability and compatibility are a substantial baseline. Keep AI generation off the ordinary keystroke path. |
| AI access | Both desktop editions announced AI answers through `=`. [Mac](https://z.weixin.qq.com/web/change-log/80), [Windows](https://z.weixin.qq.com/web/change-log/79) | A short keyboard invocation is already expected. A large editor alone is insufficient differentiation. |
| AI models | Both desktop editions announced DeepSeek-R1/Hunyuan-T1 deep answers. [Mac](https://z.weixin.qq.com/web/change-log/112), [Windows](https://z.weixin.qq.com/web/change-log/113) | Compare actual tasks; do not describe the competitor as lacking capable AI. |
| Writing and translation | Both desktop editions announced expression optimization and translation while writing through `=`. [Mac](https://z.weixin.qq.com/web/change-log/123), [Windows](https://z.weixin.qq.com/web/change-log/124) | Generic polish and translation are baseline capabilities. Improve intent control, personal vocabulary and delivery. |
| Cross-device tools | Desktop notes describe clipboard, vocabulary and phrase synchronization. [Mac](https://z.weixin.qq.com/web/change-log/48), [Windows](https://z.weixin.qq.com/web/change-log/49) | Device sync is already a competitor strength; it is a lower initial priority than the writing loop. |
| Offline mode | Both desktop editions announced standalone mode. [Mac](https://z.weixin.qq.com/web/change-log/88), [Windows](https://z.weixin.qq.com/web/change-log/87) | Do not claim privacy differentiation simply because local mode exists. Explain exactly which text a requested AI action sends. |
| Voice expansion | Desktop 3.0.0 notes announce AI recording. [Mac](https://z.weixin.qq.com/web/change-log/177), [Windows](https://z.weixin.qq.com/web/change-log/176) | Keep the user-selected keyboard direction focused; voice is not the lead scenario. |

The inspected official notes do not document bring-your-own ChatGPT sign-in, editable model prompts or per-change acceptance. Their absence from these notes is not proof that no current build has them. Performance, accuracy, resource consumption and relative privacy have not been compared experimentally. Mobile app listings and WeChat messaging-client features are not used as desktop-input-method evidence.

## Recommended features and priority

| Priority | Experience | Concrete benefit | Current shadowinput state |
| --- | --- | --- | --- |
| First | Continue with ChatGPT, list authorized models, explain plan usage and expose usage management | Eligible subscribers can start without creating an API key | OAuth, encrypted credentials, model discovery, account switching, recovery actions and usage link implemented; real account acceptance pending |
| First | Select text, invoke a compact panel, choose an action, review and apply | Reduce navigation away from the user's current editor | AX/UIA selected-range adapters, compact panel and guarded application implemented; native editor acceptance pending |
| First | Show changes beside the original; preserve names, numbers and factual commitments | Make edits easy to inspect and reduce unwanted rewriting | Diff and advisory number/date/glossary/commitment review implemented; partial acceptance and external undo remain proposed |
| First | Editable personal terms and style cards, with optional application defaults | Preserve product names and reuse preferences such as concise, natural or formal | Editable local styles, glossary and per-operation defaults implemented; per-application defaults proposed; no ChatGPT memory import |
| Next | Shortcut actions such as concise rewrite, English email, natural reply and explain changes | Turn repeated writing intentions into predictable commands | Four generic operations implemented; reusable action presets proposed |
| Later | Optional short completion near the caret, with explicit acceptance | Reduce typing while retaining control | Proposed; current continuation produces a complete draft in the editor, not inline completion |
| Later | Import/export community action templates and optional additional providers | Let the open-source community adapt the tool to specific work | Proposed; templates should be declarative and reviewable |

The first release should emphasize account connection and a reliable, inspectable editing loop. Personal style joins the inspectable writing loop; comparative quality still requires evaluation. Full IME replacement, continuous prediction, device sync, voice capture, skins and a plugin marketplace should not delay acceptance of that loop. Both operating systems remain in scope at every stage.

## Proposed primary interaction

1. The user selects a draft in a supported application and invokes shadowinput.
2. A compact panel shows the exact source and the last-used action. The normal application window remains available for longer editing and settings.
3. The user chooses a named action or types a short instruction; only an explicit generation action sends the displayed request.
4. The panel previews the result and, when implemented, highlights additions and deletions. Enter applies only a completed result while the panel owns focus; Escape cancels or dismisses it.
5. Application adapters verify the captured destination and source revision before replacement. If the target changed, offer copying the result instead of overwriting unrelated text. External undo is offered only when the adapter can verify and reverse its own transaction.

The current implementation captures a destination but pastes at its current selection; it does not yet guarantee the original selected range. Direct selection capture, exact replacement and reversible transactions require native adapter work and tests. Do not advertise universal application support before those tests.

For example, a user selects a rough Chinese email, chooses **English email**, and reviews a concise translation that preserves the stated deadline and product name. The measurable improvement is the number of steps and edits needed to send an accurate message, not the presence of a translation button.

## ChatGPT connection contract

[OpenAI's open-source/local-app flow](https://developers.openai.com/siwc/token-sharing-open-source) allows an application to request permission for eligible Responses requests to use the user's ChatGPT plan. Plan usage requires a grant; identity sign-in alone is insufficient. Eligibility, accessible models and limits depend on the authorized account. Do not promise unlimited or universally free inference.

This flow does not grant access to ChatGPT conversations or other account context. Personal style and vocabulary must come from explicit local preferences or user-provided examples. A paid or remotely hosted offering requires revisiting the corresponding OpenAI integration route instead of assuming the OSS route covers it.

Follow the [official experience guidance](https://developers.openai.com/siwc/ui-ux-guidelines): explain plan use after first authorization, show which plan is in use, link to usage management and give a useful limit-recovery action. Credentials stay in the native protected store; requests go directly from the desktop app to the authorized provider.

For optional future API-key or local providers, keep the account, endpoint, billing source and data destination explicit. Do not silently switch a failed ChatGPT request to another provider.

## Personalization and completion design

Personal style should be a visible, editable set of rules: preferred tone, length, target language and terms that should remain unchanged. Start with user-authored rules. Learning from accepted edits should be opt-in, show what was saved and allow deletion or export. Application defaults can select a style without collecting conversation history; additional context must be user-selected and visible before sending.

Optional completion should produce one short suggestion after an explicit trigger initially. A later pause-to-suggest mode requires opt-in, request cancellation and source-revision checks. Accept with Tab only while a suggestion is active and the host is not using that key for composition or navigation; otherwise preserve the host's behavior. Never insert a late response after the user edits or switches focus. Disable completion in unsupported or protected fields.

## Evaluation plan

Use matched tasks and disclose the OS, application build, available AI mode/model and network conditions for each run. Test both a complete WeType workflow and shadowinput's complete workflow; do not measure only provider response time.

| Scenario | What to observe |
| --- | --- |
| Short natural reply | Invocation steps, preservation of meaning, extra edits before acceptance |
| Chinese-to-English work email | Name/number/deadline preservation, tone, translation corrections |
| Concise rewrite | Removal of repetition without invented facts or promises |
| Product-name-heavy mixed-language draft | Term preservation and benefit of an explicit glossary |
| Cancellation and slow response | Continued typing, rejected stale results and source preservation |
| Window/selection change before applying | Wrong-target prevention and recoverable delivery |
| New account and exhausted usage | Time to first useful result, understandable account/limit handling |

Record invocation-to-first-preview and invocation-to-accepted-result latency distributions, manual interaction count, acceptance without further editing, factual-change defects and delivery success. Treat any observed wrong-target overwrite or unwanted submission as a release-blocking defect. Report results rather than claiming faster or more accurate behavior without measurements.

Cover native text editors, browser textareas/contenteditable, WeChat, a work chat client, email and a document editor on both systems. Include Chinese composition, multiline text, emoji, canceled requests, revoked permissions and an unavailable network. Windows cross-compilation does not replace native runtime acceptance.

## Implementation boundary

This study changes research and prioritization only. The source-confirmed implementation and outstanding manual tests remain owned by the [shadowinput project document](../project/shadowinput.md) and [writing specification](../../specs/011-writing.md). New behavior needs specification updates alongside implementation. Competitor documentation is evidence for workflow design, not a source of reusable proprietary code.
