package chatgptplan

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type memoryStore struct {
	session Session
	writes  int
}

func (s *memoryStore) Load() (Session, error) { return s.session, nil }
func (s *memoryStore) Save(v Session) error   { s.session = v; s.writes++; return nil }

func encode(v any) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
func signedToken(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	body := encode(map[string]string{"alg": "RS256", "kid": "test-key"}) + "." + encode(claims)
	hash := sha256.Sum256([]byte(body))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	return body + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func TestSignInValidatesIdentityAndPlanGrant(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"success", "wrong-state", "wrong-nonce", "wrong-audience", "expired", "bad-signature", "no-plan", "denied", "wrong-client"} {
		t.Run(scenario, func(t *testing.T) {
			store := &memoryStore{session: Session{HostID: "urn:uuid:54dc5213-f650-45c5-b900-6b7ebd563bd3"}}
			if scenario == "wrong-client" {
				store.session.ClientID = "oaiapp_original"
			}
			client, _ := New(store)
			var nonce, challenge string
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/.well-known/jwks.json":
					json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]string{"kty": "RSA", "kid": "test-key", "alg": "RS256", "use": "sig", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": "AQAB"}}})
				case "/api/accounts/oauth/token":
					if r.ParseForm() != nil {
						t.Error("bad form")
					}
					sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
					if base64.RawURLEncoding.EncodeToString(sum[:]) != challenge || r.Form.Get("client_id") != "oaiapp_test" || r.Form.Get("resource") != resource {
						t.Error("incorrect code exchange")
					}
					claims := map[string]any{"iss": server.URL, "sub": "subject-1", "aud": "oaiapp_test", "nonce": nonce, "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "email": "test@example.invalid"}
					if scenario == "wrong-nonce" {
						claims["nonce"] = "different"
					}
					if scenario == "wrong-audience" {
						claims["aud"] = "different"
					}
					if scenario == "expired" {
						claims["exp"] = time.Now().Add(-time.Hour).Unix()
					}
					id := signedToken(t, key, claims)
					if scenario == "bad-signature" {
						id = id[:len(id)-20] + strings.Repeat("A", 20)
					}
					scope := "openid offline_access " + planScope
					if scenario == "no-plan" {
						scope = "openid"
					}
					json.NewEncoder(w).Encode(Tokens{Access: "test-access", Refresh: "test-refresh", ID: id, Type: "Bearer", Scope: scope, ExpiresIn: 3600})
				default:
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			client.issuer = server.URL
			client.http = server.Client()
			err := client.SignIn(context.Background(), func(raw string) error {
				u, e := url.Parse(raw)
				if e != nil {
					return e
				}
				q := u.Query()
				nonce = q.Get("nonce")
				challenge = q.Get("code_challenge")
				if nonce == "" || q.Get("ext_agent_host_id") == "" || !hasScope(q.Get("scope"), planScope) {
					t.Error("missing authorization fields")
				}
				cb, _ := url.Parse(q.Get("redirect_uri"))
				if cb.Hostname() != "127.0.0.1" {
					t.Error("invalid loopback")
				}
				v := url.Values{"state": {q.Get("state")}, "code": {"code"}, "client_id": {"oaiapp_test"}}
				if scenario == "wrong-state" {
					v.Set("state", "wrong")
				}
				if scenario == "denied" {
					v.Set("error", "access_denied")
				}
				cb.RawQuery = v.Encode()
				res, e := http.Get(cb.String())
				if e == nil {
					res.Body.Close()
				}
				return e
			}, nil)
			if scenario == "success" {
				if err != nil || !store.session.Connected() {
					t.Fatalf("sign-in failed: %v", err)
				}
			} else if err == nil || store.session.Tokens.Access != "" {
				t.Fatalf("unsafe sign-in accepted: %s", scenario)
			}
		})
	}
}

func TestRefreshSerializedAndRotated(t *testing.T) {
	store := &memoryStore{session: Session{ClientID: "oaiapp_saved", Subject: "s", Tokens: Tokens{Refresh: "old", Scope: planScope}}}
	c, _ := New(store)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		r.ParseForm()
		if r.Form.Get("client_id") != "oaiapp_saved" || r.Form.Get("refresh_token") != "old" || r.Form.Has("scope") {
			t.Error("incorrect refresh grant")
		}
		json.NewEncoder(w).Encode(Tokens{Access: "renewed", Refresh: "rotated", Type: "Bearer", ExpiresIn: 3600})
	}))
	defer server.Close()
	c.issuer = server.URL
	c.http = server.Client()
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, err := c.accessToken(context.Background())
			if err != nil || token != "renewed" {
				t.Errorf("refresh: %v", err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 || store.session.Tokens.Refresh != "rotated" {
		t.Fatal("refresh rotation was replayed or not persisted")
	}
}

func TestStreamNeverDeliversIncompleteOutput(t *testing.T) {
	delta := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n"
	for _, end := range []string{"", "data: [DONE]\n\n", "data: {\"type\":\"response.incomplete\"}\n\n", "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"subscription_sharing_usage_limit_exceeded\"}}}\n\n"} {
		out, err := readStream(strings.NewReader(delta+end), nil)
		if err == nil || out != "" {
			t.Fatalf("delivered incomplete text: %q, %v", out, err)
		}
	}
	var preview string
	out, err := readStream(strings.NewReader(delta+"data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\"}}\n\n"), func(s string) { preview = s })
	if err != nil || out != "hello" || preview != out {
		t.Fatal(fmt.Sprint(out, preview, err))
	}
}
