// Package shadowinput composes a keyboard-first application from Arion adapters.
package shadowinput

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"arion.local/core/application"
	"arion.local/core/domain"
	"arion.local/core/domain/ports"
	"arion.local/desktop/infrastructure/browser"
	"arion.local/desktop/infrastructure/chatgptplan"
	"arion.local/desktop/infrastructure/clipboardos"
	"arion.local/desktop/infrastructure/llm/engine"
	"arion.local/desktop/infrastructure/localwriting"
	"arion.local/desktop/infrastructure/rewriting"
	"arion.local/desktop/ui/pages/writing"
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type app struct {
	s                 state.State
	client            *chatgptplan.Client
	win               *mygo.Window
	panel             *mygo.Window
	target            clipboardos.Target
	directory         string
	job               uint64
	cancel            context.CancelFunc
	quitting          bool
	local             *rewriting.Service
	localDone         chan struct{}
	localCancel       context.CancelFunc
	localJob          uint64
	localStatusQueued atomic.Bool
}

// Run starts no microphone, background download, upstream account or telemetry.
func Run() error {
	mygo.App.SetName(state.Name)
	if !mygo.App.RequestSingleInstanceLock() {
		return nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	a := &app{s: state.NewState(), directory: filepath.Join(base, state.Name)}
	a.loadProfile()
	a.loadPreferences()
	mygo.App.WhenReady(func() { a.ready() })
	return mygo.App.Run()
}

func (a *app) ready() {
	a.s.AccessRequired = needsAccess()
	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{{Role: mygo.RoleAppMenu}, {Role: mygo.RoleEditMenu}}))
	a.win = mygo.NewWindow(mygo.WindowOptions{Title: state.Name, Width: 820, Height: 820, MinWidth: 620, MinHeight: 640, Content: ui.View(writing.View(&a.s, a.actions()))})
	a.win.OnClose(func(e *mygo.CloseEvent) {
		if !a.quitting {
			e.PreventDefault()
			a.hide()
		}
	})
	a.win.OnFocus(func() { a.s.AccessRequired = needsAccess(); a.invalidate() })
	mygo.App.OnWindowAllClosed(func() {})
	mygo.App.OnActivate(func(bool) { a.show() })
	mygo.App.OnSecondInstance(func([]string, string) { a.show() })
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { a.quitting = true; a.stop(); a.releaseTarget(); a.closeLocal() })
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) { mygo.GlobalShortcut.UnregisterAll() })
	if err := mygo.GlobalShortcut.Register(state.Shortcut, a.capture); err != nil {
		a.s.Failure = copy.WriteShortcutError
	}
	a.installTray()
	a.initLocal()
	if a.s.Provider == state.ProviderChatGPT {
		a.initAccount()
	}

}

func (a *app) actions() state.Actions {
	return state.Actions{SignIn: a.signIn, Logout: a.logout, RefreshModels: a.models, Generate: a.generate, Cancel: a.stop,
		Import: a.importText, Copy: a.copyResult, Paste: a.pasteResult, Hide: a.hide, Changed: a.changed, Preferences: a.savePreferences,
		RequestAccess:   requestAccess,
		ProviderChanged: a.providerChanged, DownloadLocal: func() { a.startLocal(true, true) }, EnableLocal: func() { a.startLocal(false, true) }, DisableLocal: func() { a.startLocal(false, false) }, CancelLocal: func() { a.startLocal(false, false) },
		EditSource: a.editSource, TaskChanged: a.taskChanged, StyleChanged: a.styleChanged, OpenEditor: a.openEditor, OpenStyles: a.openStyles, SaveStyle: a.saveStyle, NewStyle: a.newStyle, DeleteStyle: a.deleteStyle, SaveTerms: a.saveTerms, SetDefault: a.setDefault, RetryProfile: a.loadProfile, SwitchAccount: a.switchAccount,
		ManageUsage: func() { a.open("https://chatgpt.com/settings/usage") }, Source: func() { a.open("https://github.com/simpledavid/shadowinput-desktop") }}
}

func (a *app) show() {
	if a.panel != nil {
		a.panel.Show()
		a.panel.Focus()
		a.invalidate()
		return
	}
	a.s.Compact = false
	if a.win != nil {
		a.win.Show()
		a.win.Focus()
		a.invalidate()
	}
}
func (a *app) hide() {
	if a.s.Local.Busy {
		a.startLocal(false, false)
	}
	a.stop()
	a.releaseTarget()
	a.closePanel()
	a.s.Compact = false
	a.win.Hide()
}
func (a *app) open(url string) {
	go func() {
		if browser.Open(url) != nil {
			mygo.RunOnMain(func() { a.s.Failure = copy.WriteNetworkError; a.invalidate() })
		}
	}()
}
func (a *app) releaseTarget() {
	if a.target != nil {
		a.target.Close()
		a.target = nil
	}
	a.s.HasTarget = false
}
func (a *app) changed() { a.s.InvalidateResult() }

func (a *app) stop() {
	a.job++
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	if a.s.Busy {
		a.s.InvalidateResult()
		a.s.Notice = copy.WriteCanceled
	}
	a.s.Busy = false
	if a.win != nil {
		a.invalidate()
	}
}

func (a *app) start(timeout time.Duration) (context.Context, uint64) {
	a.job++
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	a.cancel = cancel
	a.s.Busy = true
	a.s.Failure = ""
	a.s.Notice = ""
	a.invalidate()
	return ctx, a.job
}

// finish rejects late work from a canceled or replaced operation.
func (a *app) finish(job uint64, err error, apply func()) {
	mygo.RunOnMain(func() {
		if a.quitting || job != a.job {
			return
		}
		if a.cancel != nil {
			a.cancel()
			a.cancel = nil
		}
		a.s.Busy = false
		a.s.Notice = ""
		if err != nil {
			a.s.InvalidateResult()
			a.s.Failure = failureKey(err)
			if errors.Is(err, chatgptplan.ErrAuth) || errors.Is(err, chatgptplan.ErrPermission) {
				a.s.Connected = false
				a.s.Models = nil
				a.s.Model = ""
			}
		} else if apply != nil {
			apply()
		}
		a.invalidate()
	})
}

func (a *app) updateAccount() { a.s.Email, a.s.Connected = a.client.Account() }

func (a *app) signIn() {
	if a.panel != nil {
		a.openEditor()
	}
	if a.s.Busy {
		return
	}
	if a.client == nil {
		a.s.Failure = copy.WriteStoreError
		return
	}
	ctx, job := a.start(5 * time.Minute)
	a.s.Notice = copy.WriteSignInWaiting
	message := copy.For(a.s.Locale).Text(copy.WriteCallback)
	go func() {
		err := a.client.SignIn(ctx, browser.Open, func() []byte {
			return []byte("<!doctype html><meta charset=utf-8><title>shadowinput</title><p>" + html.EscapeString(message) + "</p>")
		})
		a.finish(job, err, func() { a.updateAccount(); a.models() })
	}()
}

func (a *app) logout() {
	if a.s.Busy || a.client == nil {
		return
	}
	_, job := a.start(30 * time.Second)
	go func() {
		err := a.client.Logout()
		a.finish(job, err, func() { a.updateAccount(); a.s.Models = nil; a.s.Model = ""; a.changed() })
	}()
}

func (a *app) models() {
	if a.s.Provider != state.ProviderChatGPT || a.s.Busy || a.client == nil || !a.s.Connected {
		return
	}
	ctx, job := a.start(45 * time.Second)
	go func() {
		models, err := a.client.Models(ctx)
		a.finish(job, err, func() {
			a.s.Models = nil
			a.s.ModelNames = map[string]string{}
			for _, m := range models {
				a.s.Models = append(a.s.Models, m.Slug)
				a.s.ModelNames[m.Slug] = m.Name
			}
			if a.s.ModelNames[a.s.Model] == "" {
				a.s.Model = a.s.Models[0]
				a.changed()
			}
			a.savePreferences()
		})
	}()
}

func (a *app) generate() {
	if !a.s.CanGenerate() {
		return
	}
	r := domain.WritingRequest{Task: a.s.Task, Text: a.s.Source, Instruction: a.s.Instruction, Language: a.s.Language, Model: a.s.Model, Terms: append([]string(nil), a.s.Profile.Terms...)}
	var generator ports.WritingGenerator
	if a.s.Provider == state.ProviderLocal {
		if a.local == nil || !a.local.Ready() {
			a.s.Failure = copy.WriteLocalUnavailable
			return
		}
		r.Model = rewriting.DefaultModel
		generator = localwriting.Generator{Chat: a.local}
	} else {
		if a.client == nil {
			a.s.Failure = copy.WriteStoreError
			return
		}
		generator = a.client
	}
	if style, ok := a.s.Profile.Style(a.s.StyleID); ok {
		r.Style = &style
	}
	if err := r.Validate(); err != nil {
		a.s.Failure = copy.WriteInputError
		return
	}
	a.changed()
	ctx, job := a.start(90 * time.Second)
	go func() {
		output, err := (application.Writer{Generator: generator}).Write(ctx, r, func(text string) {
			mygo.RunOnMain(func() {
				if job == a.job && !a.quitting {
					a.s.Output = text
					a.invalidate()
				}
			})
		})
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		var review domain.WritingReview
		if err == nil {
			review = domain.ReviewWriting(r.Text, output, r.Terms)
		}
		a.finish(job, err, func() {
			a.s.Output = output
			a.s.Completed = true
			a.s.Review = review
			a.s.Reviewed = false
		})
	}()
}

func failureKey(err error) copy.Key {
	switch {
	case errors.Is(err, engine.ErrInputBudget):
		return copy.WriteLocalInputBudget
	case errors.Is(err, engine.ErrOutputBudget), errors.Is(err, localwriting.ErrIncomplete):
		return copy.WriteLocalOutputBudget
	case errors.Is(err, engine.ErrControlInput):
		return copy.WriteLocalControlInput
	case domain.CodeOf(err) == domain.CodeModelMissing || domain.CodeOf(err) == domain.CodeModelUnsupported:
		return copy.WriteLocalUnavailable
	case domain.CodeOf(err) == domain.CodeCancelled:
		return copy.WriteCanceled
	case errors.Is(err, chatgptplan.ErrPermission):
		return copy.WritePermissionError
	case errors.Is(err, chatgptplan.ErrAuth):
		return copy.WriteAuthError
	case errors.Is(err, chatgptplan.ErrModel):
		return copy.WriteModelError
	case errors.Is(err, chatgptplan.ErrLimit):
		return copy.WriteLimitError
	case errors.Is(err, chatgptplan.ErrIncomplete):
		return copy.WriteIncompleteError
	case errors.Is(err, context.DeadlineExceeded):
		return copy.WriteTimeoutError
	case errors.Is(err, chatgptplan.ErrNetwork):
		return copy.WriteNetworkError
	case errors.Is(err, context.Canceled):
		return copy.WriteCanceled
	case errors.Is(err, domain.ErrWritingInput):
		return copy.WriteInputError
	default:
		return copy.WriteStoreError
	}
}

type preferences struct {
	Provider string
	Locale   copy.Locale
	Model    string
}

func (a *app) loadPreferences() {
	b, err := os.ReadFile(filepath.Join(a.directory, "preferences.json"))
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	var p preferences
	if err != nil || json.Unmarshal(b, &p) != nil || (p.Locale != copy.English && p.Locale != copy.SimplifiedChinese) {
		a.s.Failure = copy.WriteStoreError
		return
	}
	if p.Provider != "" && p.Provider != state.ProviderLocal && p.Provider != state.ProviderChatGPT {
		a.s.Failure = copy.WriteStoreError
		return
	}
	a.s.Locale, a.s.Model = p.Locale, p.Model
	a.s.Provider = p.Provider
	if a.s.Provider == "" {
		a.s.Provider = state.ProviderChatGPT
	} // Preserve the previous preview's account workflow.
}
func (a *app) savePreferences() {
	b, err := json.Marshal(preferences{Provider: a.s.Provider, Locale: a.s.Locale, Model: a.s.Model})
	if err == nil {
		err = os.MkdirAll(a.directory, 0700)
	}
	if err == nil {
		err = os.WriteFile(filepath.Join(a.directory, "preferences.json"), b, 0600)
	}
	if err != nil {
		a.s.Failure = copy.WriteStoreError
	}
}
