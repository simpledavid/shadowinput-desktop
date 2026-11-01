//go:build windows

package shadowinput

import (
	"arion.local/desktop/infrastructure/shadowime"
	"os"
	"path/filepath"
)

func Maintenance(args []string) (bool, error) {
	if len(args) != 1 || args[0] != "--remove-input-method" {
		return false, nil
	}
	executable, err := os.Executable()
	if err != nil {
		return true, err
	}
	return true, shadowime.Remove(filepath.Dir(executable))
}
