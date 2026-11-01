package shadowime

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// The helper contract rejects installing over a different registered build.
type nativeInstallation struct {
	build              string
	enabled, migration bool
	failUpdate         bool
}

func (n *nativeInstallation) run(args ...string) ([]byte, error) {
	switch args[0] {
	case "verify":
		return nil, nil
	case "enable":
		n.enabled = true
		return nil, nil
	case "query":
		return json.Marshal(map[string]any{"registered": n.build != "", "enabled": n.enabled, "build": n.build, "upgradeRequired": n.migration})
	default:
		return nil, errors.New("unexpected command")
	}
}
func (n *nativeInstallation) elevate(args []string) error {
	switch args[0] {
	case "install":
		if n.migration || n.build != "" && n.build != args[2] {
			return errors.New("native helper: revision mismatch")
		}
	case "update":
		if n.build == "" || len(args) != 4 || args[3] != n.build {
			return errors.New("native helper: previous build mismatch")
		}
		if n.failUpdate {
			return errors.New("native helper: upgrade rolled back")
		}
	default:
		return errors.New("unexpected elevated command")
	}
	n.build, n.migration = args[2], false
	return nil
}

func TestTSFInstallLifecycle(t *testing.T) {
	a, b := strings.Repeat("a", 64), strings.Repeat("b", 64)
	for _, current := range []string{"", a, b} {
		t.Run("from-"+current, func(t *testing.T) {
			native := &nativeInstallation{build: current}
			for range 2 {
				if err := installTSF("payload.dll", b, native.run, native.elevate); err != nil {
					t.Fatal(err)
				}
				if native.build != b || !native.enabled {
					t.Fatal("new build is not installed and enabled")
				}
			}
		})
	}
}
func TestTSFMigrationAndFailedUpgrade(t *testing.T) {
	a, b := strings.Repeat("a", 64), strings.Repeat("b", 64)
	native := &nativeInstallation{build: b, migration: true}
	if err := installTSF("payload.dll", b, native.run, native.elevate); err != nil {
		t.Fatal(err)
	}
	native = &nativeInstallation{build: a, failUpdate: true}
	if err := installTSF("payload.dll", b, native.run, native.elevate); err == nil || native.build != a || native.enabled {
		t.Fatal("failed upgrade was accepted or changed the previous registration")
	}
	native.failUpdate = false
	if err := installTSF("payload.dll", b, native.run, native.elevate); err != nil {
		t.Fatal(err)
	}
}

func TestTSFUnknownStateDoesNotMutateRegistration(t *testing.T) {
	for _, data := range []string{`{}`, `broken`, `{"registered":true,"enabled":true,"build":"unknown","upgradeRequired":false}`} {
		mutated := false
		run := func(args ...string) ([]byte, error) {
			if args[0] == "verify" {
				return nil, nil
			}
			if args[0] != "query" {
				mutated = true
			}
			return []byte(data), nil
		}
		err := installTSF("payload.dll", strings.Repeat("b", 64), run, func([]string) error { mutated = true; return nil })
		if err == nil || mutated {
			t.Fatal("unknown native state allowed installation")
		}
	}
}
