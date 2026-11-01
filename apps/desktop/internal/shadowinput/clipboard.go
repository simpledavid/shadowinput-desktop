package shadowinput

import (
	"arion.local/core/domain"
	"arion.local/desktop/infrastructure/clipboardos"
	"arion.local/desktop/ui/shared/copy"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"time"
)

func (a *app) capture() {
	if a.s.Busy {
		a.show()
		return
	}
	a.releaseTarget()
	a.s.AccessRequired = needsAccess()
	if !a.s.AccessRequired {
		selected := clipboardos.CaptureWritingTarget()
		if selected != nil {
			a.changed()
			a.target = selected
			a.s.HasTarget = true
			a.s.Source = selected.Text()
			a.s.Notice = copy.WriteSelectionCaptured
		}
	}
	if a.target == nil {
		a.s.Notice = copy.WriteSelectionUnavailable
	}
	a.showPanel()
}

func (a *app) importText() {
	if a.s.Busy {
		return
	}
	b, err := mygo.Clipboard.ReadFormat(transfer.Text)
	if err != nil || len(b) > 64000 {
		a.s.Failure = copy.WriteClipboardError
		return
	}
	a.releaseTarget()
	a.changed()
	a.s.Source = string(b)
	a.invalidate()
}

func (a *app) copyResult() {
	if !a.s.CanDeliver() {
		return
	}
	if err := mygo.Clipboard.Write(transfer.TextData(a.s.Output)); err != nil {
		a.s.Failure = copy.WriteClipboardError
		return
	}
	if err := mygo.Clipboard.Flush(); err != nil {
		a.s.Failure = copy.WriteClipboardError
		return
	}
	a.s.Notice = copy.WriteCopied
	a.s.Failure = ""
}

func (a *app) pasteResult() {
	if !a.s.CanDeliver() || a.target == nil {
		return
	}
	target := a.target
	a.target = nil
	a.s.HasTarget = false
	output := a.s.Output
	a.job++
	job := a.job
	a.s.Busy = true
	a.s.Failure = ""
	a.hideForDelivery()
	fail := func(key copy.Key) { target.Close(); a.s.Busy = false; a.s.Failure = key; a.show() }
	if target.Activate() != nil {
		fail(copy.WriteTargetError)
		return
	}
	deadline := time.Now().Add(700 * time.Millisecond)
	var check func()
	check = func() {
		if a.quitting || job != a.job {
			target.Close()
			return
		}
		if !target.Ready() && time.Now().Before(deadline) {
			time.AfterFunc(25*time.Millisecond, func() { mygo.RunOnMain(check) })
			return
		}
		if target.Check() != nil {
			fail(copy.WriteTargetError)
			return
		}
		unlock, ok := clipboardos.TryLock()
		if !ok {
			fail(copy.WriteClipboardError)
			return
		}
		defer unlock()
		revision, err := clipboardos.Write(domain.ClipTextPayload(output))
		if err != nil {
			fail(copy.WriteClipboardError)
			return
		}
		if target.Paste(revision) != nil {
			fail(copy.WritePasteError)
			return
		}
		target.Close()
		a.s.Busy = false
	}
	time.AfterFunc(25*time.Millisecond, func() { mygo.RunOnMain(check) })
}
