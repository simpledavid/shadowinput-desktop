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
	_ = tt.Click("Paste back")
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
	_ = tt.Click("Paste back")
	if delivered != 1 {
		t.Fatal("delivery allowed while busy")
	}
}

func TestWritingScreens(t *testing.T) {
	for _, locale := range []copy.Locale{copy.English, copy.SimplifiedChinese} {
		for _, dark := range []bool{false, true} {
			s := state.NewState()
			s.Locale = locale
			tt := ui.NewTester(func(c *ui.Context) { Editor(c, &s, state.Actions{}) }, 820, 820)
			tt.SetDark(dark)
			if !tt.HasText("shadowinput") {
				t.Fatal("brand missing")
			}
			if dir := os.Getenv("SHADOWINPUT_SCREENSHOTS"); dir != "" {
				name := string(locale) + "-light.png"
				if dark {
					name = string(locale) + "-dark.png"
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
			}
		}
	}
}
