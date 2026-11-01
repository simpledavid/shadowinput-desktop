package shadowinput

import (
	"arion.local/core/domain"
	"arion.local/core/domain/ports"
	"arion.local/desktop/infrastructure/inputcandidate"
	"arion.local/desktop/infrastructure/localwriting"
	"arion.local/desktop/infrastructure/rewriting"
	"arion.local/desktop/infrastructure/shadowime"
	state "arion.local/desktop/ui/shared/writing"
	"context"
	"github.com/egoist/mygo"
	"strings"
)

func (a *app) initInputMethod() {
	server, err := inputcandidate.Start(a.directory, a.candidate)
	if err != nil {
		a.s.IMEStatus = "AI 候选服务未启动；普通输入不受影响。"
		return
	}
	a.candidates = server
	_ = server.SetEnabled(a.s.CanSuggest())
	a.s.IMEStatus = shadowime.Status()
}
func (a *app) candidate(ctx context.Context, text string) (string, error) {
	type snapshot struct {
		request   domain.WritingRequest
		generator ports.WritingGenerator
	}
	ready := make(chan snapshot, 1)
	a.onMain(func() {
		s := snapshot{}
		if ctx.Err() == nil && !a.quitting && a.s.CanSuggest() && a.local != nil && a.local.Ready() {
			s.request = domain.WritingRequest{Task: domain.TaskCompletePhrase, Text: text, Model: rewriting.DefaultModel, Terms: append([]string(nil), a.s.Profile.Terms...)}
			if style, ok := a.s.Profile.Style(a.s.StyleID); ok {
				s.request.Style = &style
			}
			s.generator = localwriting.Generator{Chat: a.local}
		}
		ready <- s
	})
	var s snapshot
	select {
	case s = <-ready:
	case <-ctx.Done():
		return "", ctx.Err()
	}
	if s.generator == nil {
		return "", inputcandidate.ErrUnavailable
	}
	result, err := s.generator.Generate(ctx, s.request, nil)
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(result)
	if !inputcandidate.ValidResult(text, result) {
		return "", inputcandidate.ErrUnavailable
	}
	return result, nil
}
func (a *app) toggleCandidates() {
	a.s.AICandidates = !a.s.AICandidates
	if a.candidates != nil {
		a.candidates.Revoke()
	}
	a.savePreferences()
	a.invalidate()
	if a.s.AICandidates && a.s.Local.Installed && !a.s.Local.Busy && a.local != nil && !a.local.Ready() {
		a.startLocal(false, true)
	} else if !a.s.AICandidates && a.s.Provider != state.ProviderLocal {
		a.startLocal(false, false)
	}
}
func (a *app) installInputMethod() {
	if a.s.IMEInstalling {
		return
	}
	a.s.IMEInstalling = true
	a.s.IMEStatus = "正在安装输入法…"
	a.invalidate()
	resources, err := mygo.App.Path(mygo.PathResources)
	if err != nil {
		a.s.IMEInstalling = false
		a.s.IMEStatus = "无法定位输入法安装包。"
		a.invalidate()
		return
	}
	go func() {
		err := shadowime.Install(resources)
		a.onMain(func() {
			a.s.IMEInstalling = false
			if err != nil {
				a.s.IMEStatus = err.Error()
			} else {
				a.s.IMEStatus = shadowime.Status()
			}
			a.invalidate()
		})
	}()
}
