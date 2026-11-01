//go:build windows

package shadowime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func Status() string {
	return "安装后按 Win + 空格，选择 shadowinput。首次注册需要管理员权限。"
}
func helperCommand(helper string, args ...string) ([]byte, error) {
	cmd := exec.Command(helper, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Output()
}
func Install(resources string) error {
	helper := filepath.Join(resources, "ime", "arion-ime.exe")
	dll := filepath.Join(resources, "ime", "ArionTSF.dll")
	bytes, err := os.ReadFile(dll)
	if err != nil {
		return errors.New("当前包未包含系统输入法，请安装新的输入法预览版。")
	}
	digest := sha256.Sum256(bytes)
	hash := hex.EncodeToString(digest[:])
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	return installTSF(dll, hash, func(args ...string) ([]byte, error) {
		return helperCommand(helper, args...)
	}, func(args []string) error { return elevated(ctx, helper, args) })
}

func Remove(resources string) error {
	helper := filepath.Join(resources, "ime", "arion-ime.exe")
	if _, err := os.Stat(helper); os.IsNotExist(err) {
		return nil
	}
	bytes, err := helperCommand(helper, "query")
	if err != nil {
		return err
	}
	var state struct{ Registered bool }
	if json.Unmarshal(bytes, &state) != nil {
		return errors.New("无法确认输入法注册状态")
	}
	if !state.Registered {
		return nil
	}
	if _, err = helperCommand(helper, "disable"); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	return elevated(ctx, helper, []string{"remove"})
}

// Same bounded ShellExecuteEx contract as Arion's native installation adapter.
type shellExecuteInfo struct {
	Size, Mask                        uint32
	Window                            uintptr
	Verb, File, Parameters, Directory *uint16
	Show                              int32
	Instance, IDList                  uintptr
	Class                             *uint16
	ClassKey                          uintptr
	HotKey                            uint32
	Icon                              uintptr
	Process                           syscall.Handle
}

func elevated(ctx context.Context, helper string, args []string) error {
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, err := syscall.UTF16PtrFromString(helper)
	if err != nil {
		return err
	}
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = syscall.EscapeArg(arg)
	}
	parameters, err := syscall.UTF16PtrFromString(strings.Join(quoted, " "))
	if err != nil {
		return err
	}
	info := shellExecuteInfo{Mask: 0x40 | 0x100 | 0x400, Verb: verb, File: file, Parameters: parameters, Show: 0}
	info.Size = uint32(unsafe.Sizeof(info))
	ok, _, failure := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteExW").Call(uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return fmt.Errorf("未完成 Windows 输入法管理员授权：%w", failure)
	}
	if info.Process == 0 {
		return errors.New("输入法安装进程未启动")
	}
	defer syscall.CloseHandle(info.Process)
	for {
		result, err := syscall.WaitForSingleObject(info.Process, 100)
		if err != nil {
			return err
		}
		if result == syscall.WAIT_OBJECT_0 {
			var code uint32
			if err = syscall.GetExitCodeProcess(info.Process, &code); err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("输入法安装未完成（%d）；请切换到其他输入法并关闭使用它的编辑器后重试。", code)
			}
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}
