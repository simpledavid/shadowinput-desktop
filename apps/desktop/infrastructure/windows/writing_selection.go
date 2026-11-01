//go:build windows

package windows

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"
)

var writingStringLen = syscall.NewLazyDLL("oleaut32.dll").NewProc("SysStringLen")
var writingStringFree = syscall.NewLazyDLL("oleaut32.dll").NewProc("SysFreeString")

type WritingTarget struct {
	*HistoryTarget
	span *historyCOM
	text string
}

// UIA TextPattern is required. Unsupported and protected controls remain copy-only.
func CaptureWritingTarget() *WritingTarget {
	t := &WritingTarget{HistoryTarget: CaptureHistoryTarget()}
	if t.HistoryTarget == nil {
		return nil
	}
	span, text, err := t.readSelection()
	if err != nil || strings.TrimSpace(text) == "" {
		span.release()
		t.Close()
		return nil
	}
	t.span, t.text = span, text
	return t
}

func (t *WritingTarget) Text() string { return t.text }

func (t *WritingTarget) Close() {
	if t == nil {
		return
	}
	t.span.release()
	t.span = nil
	if t.HistoryTarget != nil {
		t.HistoryTarget.Close()
	}
}

func (t *WritingTarget) readSelection() (*historyCOM, string, error) {
	fail := errors.New("writing selection unavailable")
	if t == nil || t.HistoryTarget == nil || t.HistoryTarget.Check() != nil || !t.input.editable {
		return nil, "", fail
	}
	pattern, known := historyPattern(t.input.element, 10014)
	if !known || pattern == nil {
		return nil, "", fail
	}
	defer pattern.release()
	var ranges *historyCOM
	if !historyOK(pattern.call(5, uintptr(unsafe.Pointer(&ranges)))) || ranges == nil {
		ranges.release()
		return nil, "", fail
	}
	defer ranges.release()
	var count int32
	if !historyOK(ranges.call(3, uintptr(unsafe.Pointer(&count)))) || count != 1 {
		return nil, "", fail
	}
	var span *historyCOM
	if !historyOK(ranges.call(4, 0, uintptr(unsafe.Pointer(&span)))) || span == nil {
		span.release()
		return nil, "", fail
	}
	var bstr uintptr
	if !historyOK(span.call(12, 64001, uintptr(unsafe.Pointer(&bstr)))) || bstr == 0 {
		span.release()
		if bstr != 0 {
			writingStringFree.Call(bstr)
		}
		return nil, "", fail
	}
	defer writingStringFree.Call(bstr)
	length, _, _ := writingStringLen.Call(bstr)
	if length == 0 || length > 64000 {
		span.release()
		return nil, "", fail
	}
	units := unsafe.Slice((*uint16)(memory(bstr)), int(length))
	for _, unit := range units {
		if unit == 0 {
			span.release()
			return nil, "", fail
		}
	}
	text := syscall.UTF16ToString(units)
	if len(text) > 64000 || t.HistoryTarget.Check() != nil {
		span.release()
		return nil, "", fail
	}
	return span, text, nil
}

func (t *WritingTarget) Check() error {
	span, text, err := t.readSelection()
	if err != nil {
		return err
	}
	defer span.release()
	var same int32
	if t.span == nil || text != t.text || !historyOK(t.span.call(4, uintptr(unsafe.Pointer(span)), uintptr(unsafe.Pointer(&same)))) || same == 0 {
		return errors.New("writing selection changed")
	}
	return nil
}

func (t *WritingTarget) Paste(revision uint64) error {
	if err := t.Check(); err != nil {
		return err
	}
	return t.HistoryTarget.Paste(revision)
}
