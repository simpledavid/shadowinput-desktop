package chatgptplan

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type failingStore struct{ memoryStore }

func (*failingStore) Save(Session) error { return errors.New("storage unavailable") }

func TestSwitchAccountDropsAccountRegistrationOnlyAfterSaving(t *testing.T) {
	original := Session{ClientID: "oaiapp_old", HostID: "host", Subject: "subject", Email: "test@example.invalid", Tokens: Tokens{Access: "test-access", Refresh: "test-refresh", ID: "test-id", Scope: planScope}}
	store := &memoryStore{session: original}
	c, _ := New(store)
	if err := c.SwitchAccount(); err != nil {
		t.Fatal(err)
	}
	if c.session != (Session{HostID: "host"}) || store.session != c.session {
		t.Fatal("account registration was retained or host identity was lost")
	}
	broken := &failingStore{memoryStore{session: original}}
	c, _ = New(broken)
	if c.SwitchAccount() == nil || c.session != original {
		t.Fatal("failed save changed the active account")
	}
}

func TestRequestErrorsOfferDistinctRecovery(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   error
	}{
		{401, `{}`, ErrAuth}, {403, `{}`, ErrPermission}, {429, `{}`, ErrLimit},
		{404, `{"error":{"code":"model_not_found"}}`, ErrModel},
		{400, `{"error":{"code":"model_not_available"}}`, ErrModel},
		{400, `{"error":{"code":"subscription_sharing_unsupported_capability","param":"input"}}`, ErrRequest},
		{400, `{"detail":"Input must be a list"}`, ErrRequest},
		{422, `{"detail":"Invalid input"}`, ErrRequest},
		{403, `{"error":{"code":"subscription_sharing_usage_limit_exceeded"}}`, ErrLimit},
		{500, `private server detail`, ErrNetwork},
	} {
		t.Run(tc.want.Error()+http.StatusText(tc.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			c, _ := New(&memoryStore{session: Session{ClientID: "oaiapp_test", Subject: "test", Tokens: Tokens{Access: "test-access", Refresh: "test-refresh", Scope: planScope}, ExpiresAt: time.Now().Add(time.Hour)}})
			c.api, c.http = server.URL, server.Client()
			res, err := c.request(context.Background(), http.MethodGet, "/models", nil)
			if !errors.Is(err, tc.want) || res != nil {
				t.Fatalf("unexpected recovery: %v", err)
			}
		})
	}
}
