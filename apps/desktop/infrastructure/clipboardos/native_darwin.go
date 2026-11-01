//go:build darwin

package clipboardos

import "arion.local/desktop/infrastructure/macos"

func CaptureWritingTarget() WritingTarget {
	t := macos.CaptureWritingTarget()
	if t == nil {
		return nil
	}
	return t
}

func Revision() (uint64, bool) { return macos.ClipboardRevision() }
func Sequence() uint64         { return macos.ClipboardSequence() }
func TryLock() (func(), bool)  { return macos.TryClipboard() }
func CaptureTarget() Target {
	t := macos.CaptureHistoryTarget()
	if t == nil {
		return nil
	}
	return t
}
