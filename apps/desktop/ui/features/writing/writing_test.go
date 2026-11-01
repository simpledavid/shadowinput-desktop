package writing

import (
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestPartialResultCannotBeDelivered(t *testing.T) {
	s := state.NewState()
	s.Connected = true
	s.Source = "A draft"
	s.Output = "Partial result"
	s.HasTarget = true
	var delivered, generated int
	noop := func() {}
	a := state.Actions{SignIn: noop, Logout: noop, RefreshModels: noop, Generate: func() { generated++ }, Cancel: noop, Import: noop, Copy: func() { delivered++ }, Paste: func() { delivered++ }, Hide: noop, ManageUsage: noop, Source: noop, Changed: noop, Preferences: noop}
	tt := ui.NewTester(func(c *ui.Context) { Editor(c, &s, a) }, 820, 820)
	_ = tt.Click("Copy result")
	_ = tt.Click("Apply changes")
	_ = tt.Click("Generate")
	if delivered != 0 || generated != 0 {
		t.Fatal("incomplete result or missing model was accepted")
	}
	s.Completed = true
	tt.Frame()
	_ = tt.Click("Copy result")
	if delivered != 1 {
		t.Fatal("completed result was not available")
	}
	s.Busy = true
	tt.Frame()
	_ = tt.Click("Apply changes")
	if delivered != 1 {
		t.Fatal("delivery allowed while busy")
	}
}

func TestRiskReviewRequiresExplicitAcknowledgment(t *testing.T) {
	s := state.NewState()
	s.Connected, s.ProfileReady, s.HasTarget, s.Completed = true, true, true, true
	s.Source, s.Output = "Maybe 12", "Definitely 20"
	s.Review = state.Review{Edits: []state.Change{{Kind: "remove", Text: s.Source}, {Kind: "add", Text: s.Output}}, Issues: []state.Issue{{Kind: "numbers", Before: "12", After: "20"}}}
	var delivered int
	a := state.Actions{Copy: func() { delivered++ }, Paste: func() { delivered++ }}
	tt := ui.NewTester(func(c *ui.Context) { Editor(c, &s, a) }, 820, 1200)
	if err := tt.Click("Apply changes"); err != nil {
		t.Fatal(err)
	}
	if delivered != 0 {
		t.Fatal("unreviewed changes delivered")
	}
	if err := tt.Click(copy.For(s.Locale).Text(copy.WriteReviewed)); err != nil {
		t.Fatal(err)
	}
	if !s.Reviewed {
		t.Fatal("review checkbox did not update")
	}
	if err := tt.Click("Apply changes"); err != nil {
		t.Fatal(err)
	}
	if delivered != 1 {
		t.Fatal("reviewed changes were not available")
	}
	s.InvalidateResult()
	tt.Frame()
	if err := tt.Click("Apply changes"); err != nil {
		t.Fatal(err)
	}
	if delivered != 1 {
		t.Fatal("stale approval delivered a changed draft")
	}
}

func TestWritingScreens(t *testing.T) {
	for _, locale := range []copy.Locale{copy.English, copy.SimplifiedChinese} {
		for _, dark := range []bool{false, true} {
			for _, screen := range []string{"account", "compact", "review", "styles", "recovery"} {
				s := state.NewState()
				s.Locale, s.ProfileReady = locale, true
				w, h := 820, 820
				if screen != "account" {
					s.Connected, s.Email = true, "demo@example.invalid"
					s.Model, s.Models, s.ModelNames = "example-model", []string{"example-model"}, map[string]string{"example-model": "Available model"}
					s.Source = "预计周五给王小明发报价，预算 12 万元。"
				}
				if screen == "compact" || screen == "review" {
					s.Compact, s.HasTarget = true, true
					w, h = 560, 650
				}
				if screen == "review" {
					s.Completed, s.Output = true, "保证周四给王晓明发报价，预算 15 万元。"
					s.Review = state.Review{Edits: []state.Change{{Kind: "remove", Text: s.Source}, {Kind: "add", Text: s.Output}}, Issues: []state.Issue{{Kind: "numbers", Before: "12", After: "15"}, {Kind: "dates", Before: "周五", After: "周四"}, {Kind: "terms", Before: "王小明"}, {Kind: "commitments"}}}
				}
				if screen == "styles" {
					s.Settings = true
					s.StyleDraft, _ = s.Profile.Style(s.StyleID)
					s.TermsDraft = "shadowinput\n王小明"
				}
				if screen == "recovery" {
					s.Failure = copy.WriteAuthError
					s.Connected = false
				}
				tt := ui.NewTester(func(c *ui.Context) { Editor(c, &s, state.Actions{}) }, w, h)
				tt.SetDark(dark)
				if !tt.HasText("shadowinput") {
					t.Fatal("brand missing")
				}
				if dir := os.Getenv("SHADOWINPUT_SCREENSHOTS"); dir != "" {
					name := screen + "-" + string(locale) + "-light.png"
					if dark {
						name = screen + "-" + string(locale) + "-dark.png"
					}
					if err := os.MkdirAll(dir, 0700); err != nil {
						t.Fatal(err)
					}
					f, err := os.Create(filepath.Join(dir, name))
					if err != nil {
						t.Fatal(err)
					}
					if err = png.Encode(f, tt.Image()); err != nil {
						t.Fatal(err)
					}
					f.Close()
					if screen == "review" || screen == "styles" {
						tt.Scroll(float32(w-40), float32(h/2), 0, 2000)
						tt.Frame()
						f, err = os.Create(filepath.Join(dir, "scrolled-"+name))
						if err != nil {
							t.Fatal(err)
						}
						if err = png.Encode(f, tt.Image()); err != nil {
							t.Fatal(err)
						}
						f.Close()
					}
				}
			}
		}
	}
}
