package shadowinput

import (
	"os"
	"path/filepath"
	"testing"

	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
)

func TestProfileEditsSurviveRestartAndInvalidateApproval(t *testing.T) {
	a := &app{s: state.NewState(), directory: t.TempDir()}
	a.loadProfile()
	a.s.StyleDraft.Rules = "Use short sentences."
	a.s.Completed, a.s.Reviewed, a.s.Output = true, true, "An old result"
	a.saveStyle()
	if a.s.Completed || a.s.Reviewed || a.s.Output != "" || a.s.ProfileFailure != "" {
		t.Fatal("saved changes did not invalidate the previous approval")
	}
	a.s.TermsDraft = "  ShadowInput \nAda Lovelace\nShadowInput\n\n"
	a.saveTerms()
	a.setDefault("translate", "work")
	b := &app{s: state.NewState(), directory: a.directory}
	b.loadProfile()
	style, ok := b.s.Profile.Style("work")
	if !ok || style.Rules != "Use short sentences." || len(b.s.Profile.Terms) != 2 || b.s.Profile.Defaults["translate"] != "work" {
		t.Fatal("profile edits were not persisted")
	}
	b.deleteStyle()
	b.loadProfile()
	if _, ok = b.s.Profile.Style("work"); ok || b.s.Profile.Defaults["translate"] != "" {
		t.Fatal("deleting a style left dangling defaults")
	}
}

func TestInvalidProfileCannotBeSilentlyOverwritten(t *testing.T) {
	a := &app{s: state.NewState(), directory: t.TempDir()}
	path := filepath.Join(a.directory, "writing-profile.json")
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	a.loadProfile()
	a.newStyle()
	data, _ := os.ReadFile(path)
	if a.s.ProfileReady || string(data) != "broken" || a.s.ProfileFailure == "" {
		t.Fatal("invalid profile was silently replaced")
	}
}

func TestTaskDefaultsResetReviewAndSelectTranslationLanguage(t *testing.T) {
	a := &app{s: state.NewState(), directory: t.TempDir()}
	a.loadProfile()
	a.s.Completed, a.s.Reviewed = true, true
	a.s.Task = "translate"
	a.taskChanged()
	if a.s.StyleID != "email" || a.s.Language != "English" || a.s.Completed || a.s.Reviewed {
		t.Fatal("task change retained a stale review or lost its default")
	}
	if err := os.WriteFile(filepath.Join(a.directory, "preferences.json"), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	a.loadProfile()
	a.loadPreferences()
	if a.s.Failure != copy.WriteStoreError {
		t.Fatal("preferences error was hidden")
	}
}
