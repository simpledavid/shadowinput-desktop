package chatgptplan

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"arion.local/desktop/infrastructure/encryptedfile"
	"arion.local/desktop/infrastructure/secrets"
)

// Vault reuses Arion's AES-GCM file and OS-protected key adapters, with a
// separate namespace. It never reads Arion or Codex credentials.
type Vault struct{ Directory string }

func (v Vault) file(create bool) (*encryptedfile.File, error) {
	var key []byte
	var err error
	if create {
		key, err = secrets.Key(appName, "chatgpt-vault-key")
	} else {
		key, err = secrets.ExistingKey(appName, "chatgpt-vault-key")
	}
	if err != nil {
		return nil, err
	}
	return encryptedfile.New(filepath.Join(v.Directory, "chatgpt.enc"), key, "shadowinput-chatgpt-v1", 256<<10)
}

func (v Vault) Load() (Session, error) {
	if _, err := os.Lstat(filepath.Join(v.Directory, "chatgpt.enc")); err == nil {
		f, err := v.file(false)
		if err != nil {
			return Session{}, err
		}
		b, err := f.Read()
		if err != nil {
			return Session{}, err
		}
		var s Session
		if json.Unmarshal(b, &s) != nil || s.HostID == "" {
			return Session{}, ErrAuth
		}
		return s, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return Session{}, err
	}
	// Host identity contains no credentials and survives failed/canceled sign-in.
	path := filepath.Join(v.Directory, "host-id")
	if b, err := os.ReadFile(path); err == nil {
		if len(b) != 45 || string(b[:9]) != "urn:uuid:" {
			return Session{}, ErrAuth
		}
		return Session{HostID: string(b)}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return Session{}, err
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return Session{}, err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	host := fmt.Sprintf("urn:uuid:%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
	if err := os.MkdirAll(v.Directory, 0700); err != nil {
		return Session{}, err
	}
	if err := os.WriteFile(path, []byte(host), 0600); err != nil {
		return Session{}, err
	}
	return Session{HostID: host}, nil
}

func (v Vault) Save(s Session) error {
	_, err := os.Lstat(filepath.Join(v.Directory, "chatgpt.enc"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f, err := v.file(errors.Is(err, os.ErrNotExist))
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return f.Write(b)
}
