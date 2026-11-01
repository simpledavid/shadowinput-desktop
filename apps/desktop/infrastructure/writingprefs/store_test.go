package writingprefs

import (
	"os"
	"path/filepath"
	"testing"

	"arion.local/core/domain"
)

func TestProfilePersistenceAndCorruption(t *testing.T) {
	s := Store{t.TempDir()}
	p, err := s.Load()
	if err != nil || len(p.Styles) != 3 {
		t.Fatal("fresh profile", err)
	}
	p.Styles[0].Rules = "Keep my names unchanged"
	p.Terms = []string{"Alice", "shadowinput"}
	if err = s.Save(p); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.Load()
	if err != nil || loaded.Styles[0].Rules != p.Styles[0].Rules || len(loaded.Terms) != 2 {
		t.Fatal("profile did not round trip", err)
	}
	p.Defaults["polish"] = "missing"
	if s.Save(p) == nil {
		t.Fatal("invalid overwrite accepted")
	}
	loaded, err = s.Load()
	if err != nil || loaded.Defaults["polish"] != "work" {
		t.Fatal("invalid save changed valid file")
	}
	if err = os.WriteFile(filepath.Join(s.Directory, "writing-profile.json"), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Load(); err != domain.ErrWritingProfile {
		t.Fatal("corruption silently reset", err)
	}
}
