package shadowime

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

type tsfState struct {
	Registered, Enabled, UpgradeRequired bool
	Build                                string
}

func queryTSF(run func(...string) ([]byte, error)) (tsfState, error) {
	data, err := run("query")
	if err != nil {
		return tsfState{}, err
	}
	var wire struct {
		Registered, Enabled, UpgradeRequired *bool
		Build                                *string
	}
	if json.Unmarshal(data, &wire) != nil || wire.Registered == nil || wire.Enabled == nil || wire.UpgradeRequired == nil || wire.Build == nil {
		return tsfState{}, errors.New("无法确认输入法注册状态，已保留原组件。")
	}
	s := tsfState{Registered: *wire.Registered, Enabled: *wire.Enabled, UpgradeRequired: *wire.UpgradeRequired, Build: *wire.Build}
	digest, err := hex.DecodeString(s.Build)
	if s.Registered && (err != nil || len(digest) != 32) || !s.Registered && s.Build != "" {
		return tsfState{}, errors.New("无法确认已安装的输入法版本，已保留原组件。")
	}
	return s, nil
}

// installTSF orchestrates the native helper without owning OS registration itself.
func installTSF(dll, hash string, run func(...string) ([]byte, error), elevate func([]string) error) error {
	// Reuse the native helper's full keyboard/loader hash closure verification.
	if _, err := run("verify", dll, hash); err != nil {
		return errors.New("输入法组件校验失败，请重新下载安装包。")
	}
	before, err := queryTSF(run)
	if err != nil {
		return err
	}
	args := []string{"install", dll, hash}
	if before.Registered && (before.Build != hash || before.UpgradeRequired) {
		args = []string{"update", dll, hash, before.Build}
	}
	if err := elevate(args); err != nil {
		return err
	}
	if _, err := run("enable"); err != nil {
		return errors.New("输入法已安装；请在 Windows 语言设置中启用 shadowinput。")
	}
	after, err := queryTSF(run)
	if err != nil {
		return err
	}
	if !after.Registered || !after.Enabled || after.Build != hash || after.UpgradeRequired {
		return errors.New("输入法注册状态尚未确认，请重新打开语言设置查看。")
	}
	return nil
}
