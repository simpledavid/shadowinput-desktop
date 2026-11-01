package shadowinput

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"arion.local/desktop/infrastructure/chatgptplan"
	"arion.local/desktop/infrastructure/inputcandidate"
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
)

type candidateAccountStore struct{ session chatgptplan.Session }

func (s *candidateAccountStore) Load() (chatgptplan.Session, error) { return s.session, nil }
func (s *candidateAccountStore) Save(v chatgptplan.Session) error   { s.session = v; return nil }

type candidateTransport func(*http.Request) (*http.Response, error)

func (f candidateTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func candidateApp(t *testing.T) *app {
	t.Helper()
	client, err := chatgptplan.New(&candidateAccountStore{session: chatgptplan.Session{
		ClientID: "test-client", Subject: "test-subject", ExpiresAt: time.Now().Add(time.Hour),
		Tokens: chatgptplan.Tokens{Access: "test-only", Refresh: "test-only", Scope: "chatgpt.tokens.use.direct"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	a := &app{s: state.NewState(), directory: t.TempDir(), client: client, dispatch: func(fn func()) { fn() }}
	a.s.Provider, a.s.Model = state.ProviderChatGPT, "test-model"
	a.s.Connected, a.s.ProfileReady, a.s.AICandidates = true, true, true
	a.s.Models = []string{"test-model"}
	a.candidates, err = inputcandidate.Start(a.directory, a.candidate)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.candidates.Close)
	return a
}

func TestCandidateAuthFailureRevokesReadiness(t *testing.T) {
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			http.DefaultTransport = candidateTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
			})
			a := candidateApp(t)
			_, err := a.candidate(context.Background(), "明天一起讨论")
			if !errors.Is(err, chatgptplan.ErrAuth) && !errors.Is(err, chatgptplan.ErrPermission) {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.s.CanGenerate() || a.s.Connected || len(a.s.Models) != 0 || a.s.Model != "" {
				t.Fatal("rejected account is still presented as usable")
			}
			if _, err := os.Stat(filepath.Join(a.directory, "candidate.endpoint")); !os.IsNotExist(err) {
				t.Fatal("native candidate discovery remains enabled after authorization failure")
			}
			if a.s.Failure != copy.WriteAuthError && a.s.Failure != copy.WritePermissionError {
				t.Fatal("reauthorization reason missing")
			}
		})
	}
}

func TestCanceledCandidateCannotOverwriteCurrentProvider(t *testing.T) {
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	http.DefaultTransport = candidateTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
	})
	a := candidateApp(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	a.dispatch = func(fn func()) {
		calls++
		if calls == 2 {
			cancel()
			a.s.Provider, a.s.Failure = state.ProviderLocal, copy.WriteLocalUnavailable
		}
		fn()
	}
	_, _ = a.candidate(ctx, "明天一起讨论")
	if a.s.Failure != copy.WriteLocalUnavailable {
		t.Fatal("late account error overwrote the new provider")
	}
}
