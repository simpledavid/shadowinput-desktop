package shadowinput

import (
	"os"
	"path/filepath"
	"testing"

	view "arion.local/desktop/ui/features/writing"
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
)

// Both entry points must explain why a profile failure prevents inference.
func TestInputPageProfileFailure(t *testing.T) {
	for _, inputPage := range []bool{false, true} {
		name := "writing-editor-control"
		if inputPage {
			name = "input-method-settings"
		}
		t.Run(name, func(t *testing.T) {
			a := &app{s: state.NewState(), directory: t.TempDir()}
			a.s.InputPage, a.s.AICandidates = inputPage, true
			a.s.Local.State = "ready"
			if err := os.WriteFile(filepath.Join(a.directory, "writing-profile.json"), []byte("invalid-profile"), 0600); err != nil {
				t.Fatal(err)
			}
			a.loadProfile()
			if a.s.ProfileFailure != copy.WriteProfileError || a.s.CanGenerate() {
				t.Fatal("fixture did not reach the real profile failure gate")
			}
			tester := ui.NewTester(func(c *ui.Context) { view.Editor(c, &a.s, a.actions()) }, 820, 1200)
			if !tester.HasText(copy.For(a.s.Locale).Text(copy.WriteProfileError)) {
				t.Fatal("AI generation is blocked by profile loading, but the active page hides the reason")
			}
			if err := os.Remove(filepath.Join(a.directory, "writing-profile.json")); err != nil {
				t.Fatal(err)
			}
			if err := tester.Click(copy.For(a.s.Locale).Text(copy.WriteRetryProfile)); err != nil {
				t.Fatal(err)
			}
			if !a.s.CanGenerate() || a.s.ProfileFailure != "" {
				t.Fatal("retry did not restore AI readiness")
			}
		})
	}
}
