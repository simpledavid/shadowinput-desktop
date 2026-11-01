//go:build darwin

package macos

/*
#include <stdlib.h>
#include "native.h"
*/
import "C"

import (
	"errors"
	"strings"
	"unsafe"
)

type WritingTarget struct {
	*HistoryTarget
	text          string
	start, length int64
}

func CaptureWritingTarget() *WritingTarget {
	t := &WritingTarget{HistoryTarget: CaptureHistoryTarget()}
	if t.HistoryTarget == nil {
		return nil
	}
	text, start, length, ok := t.readSelection()
	if !ok || strings.TrimSpace(text) == "" {
		t.Close()
		return nil
	}
	t.text, t.start, t.length = text, start, length
	return t
}

func (t *WritingTarget) Text() string { return t.text }

func (t *WritingTarget) readSelection() (string, int64, int64, bool) {
	if t == nil || t.HistoryTarget == nil || t.HistoryTarget.Check() != nil {
		return "", 0, 0, false
	}
	var start, length C.longlong
	text := C.arion_writing_selection(C.int(t.pid), t.element, &start, &length)
	if text == nil {
		return "", 0, 0, false
	}
	defer C.free(unsafe.Pointer(text))
	return C.GoString(text), int64(start), int64(length), true
}

func (t *WritingTarget) Check() error {
	text, start, length, ok := t.readSelection()
	if !ok || text != t.text || start != t.start || length != t.length {
		return errors.New("writing selection changed or unavailable")
	}
	return nil
}

func (t *WritingTarget) Paste(revision uint64) error {
	if err := t.Check(); err != nil {
		return err
	}
	return t.HistoryTarget.Paste(revision)
}
