package writing

import (
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPartialResultCannotBeDelivered(t *testing.T) {
	s := state.NewState()
	s.Provider = state.ProviderChatGPT
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
	s.Provider = state.ProviderChatGPT
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
			for _, screen := range []string{"account", "compact", "review", "styles", "recovery", "local-off", "local-progress", "local-ready", "local-error", "local-compact"} {
				s := state.NewState()
				s.Provider = state.ProviderChatGPT
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
				if strings.HasPrefix(screen, "local-") {
					s.Provider = state.ProviderLocal
					s.Local = state.LocalStatus{Name: "Qwen3.5 2B · Q4_K_M", State: "off", SizeBytes: 1280835840}
					s.Connected = false
					s.Model = ""
					switch screen {
					case "local-progress":
						s.Local.State = "provisioning"
						s.Local.Phase = "downloading"
						s.Local.Busy = true
						s.Local.Done = 650000000
						s.Local.Total = 1280835840
					case "local-ready":
						s.Local.State = "ready"
						s.Local.Installed = true
					case "local-error":
						s.Local.State = "failed"
						s.Local.Failure = copy.WriteLocalDownloadError
					case "local-compact":
						s.Local.State = "ready"
						s.Local.Installed = true
						s.Compact = true
						w, h = 560, 650
					}
				}
				tt := ui.NewTester(func(c *ui.Context) { c.DisableAnimations(); Editor(c, &s, state.Actions{}) }, w, h)
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

func TestLocalSetupAndGenerationActions(t *testing.T) {
	s := state.NewState()
	s.ProfileReady = true
	s.Source = "A draft"
	s.Local = state.LocalStatus{State: "off", Name: "Qwen3.5 2B", SizeBytes: 1280835840}
	downloads, cancels, generated, stops := 0, 0, 0, 0
	a := state.Actions{DownloadLocal: func() { downloads++ }, CancelLocal: func() { cancels++ }, Generate: func() { generated++ }, DisableLocal: func() { stops++ }}
	tt := ui.NewTester(func(c *ui.Context) { Editor(c, &s, a) }, 820, 820)
	if tt.HasText("Sign in with ChatGPT") {
		t.Fatal("local setup requires account")
	}
	if err := tt.Click("Download & start"); err != nil {
		t.Fatal(err)
	}
	_ = tt.Click("Generate")
	if downloads != 1 || generated != 0 {
		t.Fatal("download or readiness gate failed")
	}
	s.Local.Busy = true
	s.Local.State = "provisioning"
	s.Local.Phase = "downloading"
	tt.Frame()
	if err := tt.Click("Cancel setup"); err != nil {
		t.Fatal(err)
	}
	if cancels != 1 {
		t.Fatal("cancel disconnected")
	}
	s.Local.Busy = false
	s.Local.State = "ready"
	s.Local.Installed = true
	tt.Frame()
	if err := tt.Click("Generate"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Stop model"); err != nil {
		t.Fatal(err)
	}
	if generated != 1 || stops != 1 {
		t.Fatal("local ready actions disconnected")
	}
}
