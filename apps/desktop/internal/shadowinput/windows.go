package shadowinput

import (
	"arion.local/desktop/ui/pages/writing"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func (a *app) invalidate() {
	if a.win != nil {
		a.win.Invalidate()
	}
	if a.panel != nil {
		a.panel.Invalidate()
	}
}
func (a *app) closePanel() {
	if a.panel != nil {
		panel := a.panel
		a.panel = nil
		panel.Close()
	}
}
func (a *app) openEditor() { a.closePanel(); a.s.Compact = false; a.s.Settings = false; a.show() }
func (a *app) openStyles() {
	if a.s.Busy {
		return
	}
	a.closePanel()
	a.s.Compact = false
	a.s.Settings = true
	a.s.StyleDraft, _ = a.s.Profile.Style(a.s.StyleID)
	a.s.TermsDraft = ""
	for i, term := range a.s.Profile.Terms {
		if i > 0 {
			a.s.TermsDraft += "\n"
		}
		a.s.TermsDraft += term
	}
	a.show()
}
func (a *app) showPanel() {
	a.closePanel()
	a.win.Hide()
	a.s.Compact = true
	a.s.Settings = false
	point := mygo.Screen.CursorScreenPoint()
	anchor := mygo.Rectangle{X: point.X, Y: point.Y}
	if a.target != nil {
		if x, y, w, h, ok := a.target.Anchor(); ok {
			anchor = mygo.Rectangle{X: x, Y: y, Width: w, Height: h}
		}
	}
	panel := mygo.NewFlyout(mygo.FlyoutOptions{ScreenAnchor: &anchor, Placement: mygo.PlacementBottomStart, Gap: 6, Width: 560, Height: 650, Hidden: true, Focusable: true, Shadow: true, Content: ui.View(writing.View(&a.s, a.actions()))})
	a.panel = panel
	panel.OnClosed(func() {
		if a.panel == panel {
			a.panel = nil
			a.s.Compact = false
			a.stop()
			a.releaseTarget()
		}
	})
	panel.Show()
	panel.Focus()
}
func (a *app) hideForDelivery() { a.closePanel(); a.s.Compact = false; a.win.Hide() }
