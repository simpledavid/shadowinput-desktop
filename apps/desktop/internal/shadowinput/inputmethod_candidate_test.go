package shadowinput

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"arion.local/desktop/infrastructure/chatgptplan"
	"arion.local/desktop/infrastructure/inputcandidate"
	"arion.local/desktop/infrastructure/llm/engine"
	"arion.local/desktop/infrastructure/llm/prompt"
	"arion.local/desktop/infrastructure/rewriting"
	state "arion.local/desktop/ui/shared/writing"
)

type candidateAccountStore struct{ session chatgptplan.Session }

func (s *candidateAccountStore) Load() (chatgptplan.Session, error) { return s.session, nil }
func (s *candidateAccountStore) Save(v chatgptplan.Session) error   { s.session = v; return nil }

type candidateTransport func(*http.Request) (*http.Response, error)

func (f candidateTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAutomaticCandidateNeverFallsBackToChatGPT(t *testing.T) {
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	calls := 0
	http.DefaultTransport = candidateTransport(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected cloud request")
	})
	client, err := chatgptplan.New(&candidateAccountStore{session: chatgptplan.Session{
		ClientID: "test-client", Subject: "test-subject", ExpiresAt: time.Now().Add(time.Hour),
		Tokens: chatgptplan.Tokens{Access: "test-only", Refresh: "test-only", Scope: "chatgpt.tokens.use.direct"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	a := &app{s: state.NewState(), client: client, dispatch: func(fn func()) { fn() }}
	a.s.Provider, a.s.Model = state.ProviderChatGPT, "test-model"
	a.s.Connected, a.s.ProfileReady, a.s.AICandidates = true, true, true
	_, err = a.candidate(context.Background(), "明天我们")
	if calls != 0 || !errors.Is(err, inputcandidate.ErrUnavailable) {
		t.Fatalf("automatic input reached cloud: calls=%d err=%v", calls, err)
	}
	if !a.s.Connected {
		t.Fatal("local candidate changed the writing account")
	}
}

type candidateLocal struct {
	ready            bool
	calls, downloads int
	request          prompt.ChatRequest
}

func (l *candidateLocal) Complete(_ context.Context, r prompt.ChatRequest) (engine.Result, error) {
	l.calls++
	l.request = r
	return engine.Result{Text: "一起讨论"}, nil
}
func (l *candidateLocal) Ready() bool { return l.ready }
func (l *candidateLocal) Status() rewriting.Status {
	if l.ready {
		return rewriting.Status{State: rewriting.Ready, Installed: true}
	}
	return rewriting.Status{State: rewriting.Off}
}
func (l *candidateLocal) OnStatus(func(rewriting.Status)) {}
func (l *candidateLocal) Refresh()                        {}
func (l *candidateLocal) Enable(context.Context) error    { l.ready = true; return nil }
func (l *candidateLocal) Download(context.Context) error  { l.downloads++; return nil }
func (l *candidateLocal) Disable()                        { l.ready = false }

func TestLocalCandidateIndependentOfWritingAccount(t *testing.T) {
	local := &candidateLocal{ready: true}
	a := &app{s: state.NewState(), local: local, dispatch: func(fn func()) { fn() }}
	a.s.Provider = state.ProviderChatGPT // Logged out; the writing account must not gate local input.
	a.s.ProfileReady, a.s.AICandidates = true, true
	a.applyLocalStatus()
	got, err := a.candidate(context.Background(), "明天我们")
	if err != nil || got != "一起讨论" || local.calls != 1 || local.downloads != 0 {
		t.Fatalf("%q %v calls=%d", got, err, local.calls)
	}
	if local.request.MaxTokens > 64 || local.request.Messages[1].Content != "明天我们" {
		t.Fatal("wrong local continuation request")
	}
	a.s.Busy = true
	if _, err = a.candidate(context.Background(), "明天我们"); !errors.Is(err, inputcandidate.ErrUnavailable) {
		t.Fatal("candidate interrupted explicit writing")
	}
	a.s.Busy = false
	a.s.AICandidates = false
	if _, err = a.candidate(context.Background(), "明天我们"); !errors.Is(err, inputcandidate.ErrUnavailable) || local.calls != 1 {
		t.Fatal("opt-out reached model")
	}
}

func TestEnableCandidatesNeverDownloadsMissingModel(t *testing.T) {
	local := &candidateLocal{}
	a := &app{s: state.NewState(), local: local, directory: t.TempDir(), dispatch: func(fn func()) { fn() }}
	a.toggleCandidates()
	if local.downloads != 0 || local.Ready() {
		t.Fatal("opt-in downloaded or started a missing model")
	}
}
