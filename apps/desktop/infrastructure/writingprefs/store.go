// Package writingprefs owns validated local writing preferences.
package writingprefs

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"arion.local/core/domain"
)

type Store struct{ Directory string }

func (s Store) Load() (domain.WritingProfile, error) {
	f, err := os.Open(filepath.Join(s.Directory, "writing-profile.json"))
	if errors.Is(err, os.ErrNotExist) {
		return domain.DefaultWritingProfile(), nil
	}
	var p domain.WritingProfile
	if err != nil {
		return p, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 256001))
	if err != nil {
		return p, err
	}
	if len(data) > 256000 || json.Unmarshal(data, &p) != nil || p.Validate() != nil {
		return p, domain.ErrWritingProfile
	}
	return p, nil
}

func (s Store) Save(p domain.WritingProfile) error {
	if err := p.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(s.Directory, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(s.Directory, ".writing-profile-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(s.Directory, "writing-profile.json"))
}
