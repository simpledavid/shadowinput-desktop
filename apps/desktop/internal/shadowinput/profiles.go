package shadowinput

import (
	"arion.local/core/domain"
	"arion.local/desktop/infrastructure/writingprefs"
	"arion.local/desktop/ui/shared/copy"
	"crypto/rand"
	"maps"
	"slices"
	"strings"
)

func (a *app) loadProfile() {
	p, err := (writingprefs.Store{Directory: a.directory}).Load()
	if err != nil {
		a.s.ProfileReady = false
		a.s.ProfileFailure = copy.WriteProfileError
		return
	}
	a.s.Profile = p
	a.s.ProfileReady = true
	a.s.ProfileFailure = ""
	a.s.StyleID = p.Defaults[a.s.Task]
	a.s.TermsDraft = strings.Join(p.Terms, "\n")
	a.styleChanged()
}
func (a *app) profileCopy() domain.WritingProfile {
	p := a.s.Profile
	p.Styles = slices.Clone(p.Styles)
	p.Terms = slices.Clone(p.Terms)
	p.Defaults = maps.Clone(p.Defaults)
	if p.Defaults == nil {
		p.Defaults = map[string]string{}
	}
	return p
}
func (a *app) saveProfile(p domain.WritingProfile) bool {
	if !a.s.ProfileReady || (writingprefs.Store{Directory: a.directory}).Save(p) != nil {
		a.s.ProfileFailure = copy.WriteProfileSaveError
		return false
	}
	a.s.Profile = p
	a.changed()
	a.s.ProfileFailure = ""
	a.s.Notice = copy.WriteProfileSaved
	return true
}
func (a *app) taskChanged() { a.s.StyleID = a.s.Profile.Defaults[a.s.Task]; a.styleChanged() }
func (a *app) styleChanged() {
	a.changed()
	a.s.StyleDraft, _ = a.s.Profile.Style(a.s.StyleID)
	if a.s.Task == "translate" && a.s.StyleDraft.Language != "" {
		a.s.Language = a.s.StyleDraft.Language
	}
}
func (a *app) saveStyle() {
	p := a.profileCopy()
	for i := range p.Styles {
		if p.Styles[i].ID == a.s.StyleID && a.s.StyleDraft.ID == a.s.StyleID {
			p.Styles[i] = a.s.StyleDraft
			a.saveProfile(p)
			return
		}
	}
	a.s.ProfileFailure = copy.WriteProfileSaveError
}
func (a *app) newStyle() {
	if a.s.Busy {
		return
	}
	p := a.profileCopy()
	s := domain.WritingStyle{ID: "style-" + rand.Text(), Name: copy.For(a.s.Locale).Text(copy.WriteNewStyleName), Tone: "natural", Length: "keep"}
	p.Styles = append(p.Styles, s)
	if a.saveProfile(p) {
		a.s.StyleID = s.ID
		a.s.StyleDraft = s
	}
}
func (a *app) deleteStyle() {
	if a.s.Busy || a.s.StyleID == "" {
		return
	}
	p := a.profileCopy()
	id := a.s.StyleID
	p.Styles = slices.DeleteFunc(p.Styles, func(s domain.WritingStyle) bool { return s.ID == id })
	for task, style := range p.Defaults {
		if style == id {
			p.Defaults[task] = ""
		}
	}
	if a.saveProfile(p) {
		a.s.StyleID = ""
		a.s.StyleDraft = domain.WritingStyle{}
	}
}
func (a *app) saveTerms() {
	p := a.profileCopy()
	p.Terms = nil
	seen := map[string]bool{}
	for _, line := range strings.Split(a.s.TermsDraft, "\n") {
		term := strings.TrimSpace(line)
		if term != "" && !seen[term] {
			p.Terms = append(p.Terms, term)
			seen[term] = true
		}
	}
	if a.saveProfile(p) {
		a.s.TermsDraft = strings.Join(p.Terms, "\n")
	}
}
func (a *app) setDefault(task, style string) {
	p := a.profileCopy()
	p.Defaults[task] = style
	a.saveProfile(p)
}
func (a *app) editSource() { a.releaseTarget(); a.changed() }
