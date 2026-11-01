//go:build darwin

package shadowime

/*
#cgo LDFLAGS: -framework Carbon -framework Cocoa
#include <stdlib.h>
int shadow_ime_register(const char *path);
*/
import "C"
import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var installMu sync.Mutex

type observation struct {
	Complete, Available, Selected, ParentFound, ParentEnabled bool
	OwnCount                                                  int
}

func installedPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library/Input Methods/ShadowInput.app")
}

// TIS registration can change only the caller's cache. A fresh process must
// confirm persistence before the UI may claim the input method is available.
func observe() (observation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	bytes, err := exec.CommandContext(ctx, filepath.Join(installedPath(), "Contents/MacOS/ArionInput"), "--arion-source-maintenance", "observe").Output()
	var result struct {
		Version     int
		Observation observation
	}
	if err == nil {
		err = json.Unmarshal(bytes, &result)
	}
	if err == nil && (result.Version != 1 || !result.Observation.Complete) {
		err = errors.New("输入法状态未知")
	}
	return result.Observation, err
}

func observedStatus(s observation) string {
	if s.Complete && s.Available && s.ParentFound && s.ParentEnabled {
		if s.Selected {
			return "已选中 shadowinput，可以直接打字。"
		}
		return "已启用。请在菜单栏输入菜单切换到 shadowinput。"
	}
	if s.Complete && s.OwnCount > 0 {
		return "系统已发现组件，但尚未启用。请在系统设置 → 键盘 → 输入法中添加 shadowinput。"
	}
	return "组件已安装，系统尚未启用。保存工作后可注销并重新登录，再检查输入法列表；目前还不能切换使用。"
}

func Status() string {
	if _, err := os.Stat(installedPath()); os.IsNotExist(err) {
		return "尚未安装系统输入法。"
	}
	s, err := observe()
	if err != nil {
		return "无法确认系统输入法状态。请查看系统设置 → 键盘 → 输入法；不要重复安装。"
	}
	return observedStatus(s)
}
func Install(resources string) error {
	installMu.Lock()
	defer installMu.Unlock()
	before, observationErr := observe()
	if before.Selected {
		return errors.New("请先切换到其他输入法，再安装或更新 shadowinput。")
	}
	if _, err := os.Stat(installedPath()); err == nil {
		if observationErr != nil {
			return errors.New("无法确认输入法使用状态，已保留原组件。请注销并重新登录后重试。")
		}
		processes, err := exec.Command("/bin/ps", "-axo", "comm=").Output()
		if err != nil {
			return errors.New("无法确认输入法进程状态，已保留原组件。")
		}
		for _, line := range strings.Split(string(processes), "\n") {
			if strings.TrimSpace(line) == filepath.Join(installedPath(), "Contents/MacOS/ArionInput") {
				return errors.New("输入法组件正在运行，已保留原版本。请切换到其他输入法，保存工作并注销、重新登录后再更新。")
			}
		}
	}
	source := filepath.Join(resources, "ShadowInput.app")
	if _, err := os.Stat(source); err != nil {
		return errors.New("当前包未包含系统输入法，请安装新的输入法预览版。")
	}
	if err := exec.Command("/usr/bin/codesign", "--verify", "--deep", "--strict", source).Run(); err != nil {
		return errors.New("输入法包校验失败，请重新下载安装包。")
	}
	id, err := exec.Command("/usr/libexec/PlistBuddy", "-c", "Print :CFBundleIdentifier", filepath.Join(source, "Contents/Info.plist")).Output()
	if err != nil || string(id) != "org.shadowinput.inputmethod.shadowinput\n" {
		return errors.New("输入法包身份不匹配。")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	parent := filepath.Join(home, "Library/Input Methods")
	if err = os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, ".shadowinput-install-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	next := filepath.Join(stage, "ShadowInput.app")
	if err = exec.Command("/usr/bin/ditto", source, next).Run(); err != nil {
		return err
	}
	target := filepath.Join(parent, "ShadowInput.app")
	backup := filepath.Join(stage, "previous.app")
	had := false
	if _, err = os.Lstat(target); err == nil {
		if err = os.Rename(target, backup); err != nil {
			return err
		}
		had = true
	}
	if err = os.Rename(next, target); err != nil {
		if had {
			os.Rename(backup, target)
		}
		return err
	}
	path := C.CString(target)
	defer C.free(unsafe.Pointer(path))
	if C.shadow_ime_register(path) != 0 {
		// Keep the installed complete bundle; first-time TIS discovery may need login.
		return errors.New("输入法已复制；请在系统设置 → 键盘 → 输入法中添加 shadowinput。若未出现，请注销后重新登录。")
	}
	after, err := observe()
	if err != nil || !after.Available || !after.ParentFound || !after.ParentEnabled {
		return errors.New("组件已安装，但系统尚未启用。请完成系统首次允许提示，再检查系统设置 → 键盘 → 输入法；若仍未出现，可保存工作后注销并重新登录。")
	}
	return nil
}
