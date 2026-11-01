# 012: shadowinput local input phrases

This derivative extends the native Pinyin entry point in [ADR 036](../docs/decisions/036-shadowinput-runtime.md). Explicit long-form writing remains covered by [011](011-writing.md). Runtime acceptance and measured limits live in the [input-method guide](../docs/project/shadowinput-ime.md).

| ID | Observable requirement | Verification |
| --- | --- | --- |
| NPH-01 | Space/numbers commit ordinary words without waiting for a model. Only committed text starts a short continuation request. | Native Rime/NSTextView fixture |
| NPH-02 | Tab/click appends once at the unchanged caret. Continued typing, deletion, navigation, focus changes and opt-out revoke old work. | Native fixture, portable request-state and broker cancellation tests; real OS acceptance separately |
| NPH-03 | Context contains only this input method's recent contiguous commits, is bounded and expires. No surrounding document reads or input history. | Shared UTF-8 context fixture and source inspection |
| NPH-04 | Automatic phrases use the local model regardless of writing-provider selection. No cloud fallback or implicit model download. | Local routing, opt-out and missing-model tests |
| NPH-05 | Complete short suffixes only; reject echo, multiline/control output, repeated boundaries and obvious assistant replies. | Generator and broker tests; filtering is heuristic |
| NPH-06 | Shadowinput defaults to existing Rime typo correction while preserving exact Pinyin candidates. | Real native engine fixture; Windows cross-build with shared default |
| NPH-07 | Input settings expose model setup and opt-in independently of ChatGPT login; profile failures retain recovery. | Production-view action and recovery tests; localized renders |

No claim of Windows runtime or editor-wide compatibility follows from a cross-build. No claim of full semantic preservation follows from surface filters.
