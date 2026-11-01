package shadowinput

import (
	"arion.local/core/domain"
	"arion.local/core/domain/ports"
	"arion.local/desktop/infrastructure/chatgptplan"
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
	_ = server.SetEnabled(a.s.AICandidates && a.s.CanGenerate())
	a.s.IMEStatus = shadowime.Status()
}
func (a *app) candidate(ctx context.Context, text string) (string, error) {
	type snapshot struct {
		request   domain.WritingRequest
		generator ports.WritingGenerator
		provider  string
		client    *chatgptplan.Client
		job       uint64
	}
	ready := make(chan snapshot, 1)
	a.onMain(func() {
		s := snapshot{provider: a.s.Provider, client: a.client, job: a.job}
		if ctx.Err() == nil && !a.quitting && a.s.AICandidates && a.s.CanGenerate() {
			s.request = domain.WritingRequest{Task: "polish", Text: text, Model: a.s.Model, Terms: append([]string(nil), a.s.Profile.Terms...), Instruction: "这是输入法当前尚未上屏的短句。给出一个更自然、简洁的同义候选；严格保留原意，不回答问题，不添加事实、承诺、解释或引号，只输出一行候选。"}
			if style, ok := a.s.Profile.Style(a.s.StyleID); ok {
				s.request.Style = &style
			}
			if a.s.Provider == state.ProviderLocal && a.local != nil && a.local.Ready() {
				s.generator = localwriting.Generator{Chat: a.local}
				s.request.Model = rewriting.DefaultModel
			} else if a.s.Provider == state.ProviderChatGPT && a.client != nil {
				s.generator = a.client
			}
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
		if ctx.Err() == nil {
			a.onMain(func() {
				if !a.quitting && ctx.Err() == nil && a.job == s.job && a.s.Provider == s.provider && a.client == s.client {
					a.setFailure(err)
					a.invalidate()
				}
			})
		}
		return "", err
	}
	result = strings.TrimSpace(result)
	if len(domain.ReviewWriting(s.request.Text, result, s.request.Terms).Issues) > 0 {
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
