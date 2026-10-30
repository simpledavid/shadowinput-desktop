// Package chatgptplan implements the documented Sign in with ChatGPT OSS flow.
package chatgptplan

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"arion.local/desktop/internal/oauthloopback"
	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	issuer    = "https://auth.openai.com"
	resource  = "https://api.openai.com/v1"
	planScope = "chatgpt.tokens.use.direct"
	appName   = "shadowinput"
)

var (
	ErrAuth       = errors.New("chatgpt authentication failed")
	ErrPermission = errors.New("chatgpt plan permission required")
	ErrNetwork    = errors.New("chatgpt request failed")
	ErrIncomplete = errors.New("chatgpt response did not complete")
	ErrLimit      = errors.New("chatgpt usage limit reached")
)

type Tokens struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ID        string `json:"id_token"`
	Type      string `json:"token_type"`
	Scope     string `json:"scope"`
	ExpiresIn int64  `json:"expires_in"`
}

type Session struct {
	ClientID, HostID, Subject, Email string
	Tokens                           Tokens
	ExpiresAt                        time.Time
}

func (s Session) Connected() bool {
	return s.ClientID != "" && s.Subject != "" && s.Tokens.Refresh != "" && hasScope(s.Tokens.Scope, planScope)
}

type Store interface {
	Load() (Session, error)
	Save(Session) error
}

type Client struct {
	mu          sync.Mutex
	store       Store
	session     Session
	http        *http.Client
	issuer, api string
}

func New(store Store) (*Client, error) {
	s, err := store.Load()
	if err != nil {
		return nil, err
	}
	return &Client{store: store, session: s, issuer: issuer, api: resource,
		http: &http.Client{Timeout: 90 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

// Account exposes display metadata only, never bearer credentials.
func (c *Client) Account() (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session.Email, c.session.Connected()
}

func (c *Client) SignIn(ctx context.Context, open func(string) error, page func() []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	verifier, challenge, err := oauthloopback.NewPKCE()
	if err != nil {
		return err
	}
	state, err := oauthloopback.NewState()
	if err != nil {
		return err
	}
	nonce, err := oauthloopback.NewState()
	if err != nil {
		return err
	}
	server, redirect, err := oauthloopback.Start(ctx, "127.0.0.1:0", "/auth/callback", page)
	if err != nil {
		return ErrAuth
	}
	defer server.Close()
	s := c.session
	clientID := s.ClientID
	if clientID == "" {
		clientID = "dynamic_agent_client"
	}
	q := url.Values{"client_id": {clientID}, "ext_agent_host_id": {s.HostID}, "response_type": {"code"},
		"redirect_uri": {redirect}, "scope": {"openid profile email offline_access resource.invoke " + planScope},
		"resource": {resource}, "state": {state}, "nonce": {nonce}, "code_challenge_method": {"S256"}, "code_challenge": {challenge}}
	if s.ClientID == "" {
		q.Set("agent_name_hint", appName)
	} else if s.Tokens.ID != "" {
		q.Set("id_token_hint", s.Tokens.ID)
	}
	if err = open(c.issuer + "/api/accounts/authorize?" + q.Encode()); err != nil {
		return ErrAuth
	}
	cb, err := server.Wait(ctx)
	if err != nil {
		return err
	}
	if !oauthloopback.MatchState(state, cb.State) || cb.Error != "" || cb.Code == "" {
		return ErrAuth
	}
	if s.ClientID == "" {
		if cb.ClientID == "" || cb.ClientID == "dynamic_agent_client" {
			return ErrAuth
		}
		s.ClientID = cb.ClientID
		// Retain the issued registration if exchange fails or the code expires.
		if err = c.store.Save(s); err != nil {
			return err
		}
		c.session = s
	} else if cb.ClientID != "" && cb.ClientID != s.ClientID {
		return ErrAuth
	}
	tok, err := c.exchange(ctx, url.Values{"grant_type": {"authorization_code"}, "client_id": {s.ClientID},
		"code": {cb.Code}, "code_verifier": {verifier}, "redirect_uri": {redirect}, "resource": {resource}})
	if err != nil {
		return err
	}
	verifiedContext := oidc.ClientContext(ctx, c.http)
	keys := oidc.NewRemoteKeySet(verifiedContext, c.issuer+"/.well-known/jwks.json")
	id, err := oidc.NewVerifier(c.issuer, keys, &oidc.Config{ClientID: s.ClientID}).Verify(verifiedContext, tok.ID)
	if err != nil || id.Nonce != nonce || id.Subject == "" {
		return ErrAuth
	}
	if s.Subject != "" && s.Subject != id.Subject {
		return ErrAuth
	}
	var claims struct {
		Email string `json:"email"`
	}
	if err = id.Claims(&claims); err != nil {
		return ErrAuth
	}
	if !hasScope(tok.Scope, planScope) {
		return ErrPermission
	}
	if tok.Refresh == "" {
		return ErrAuth
	}
	s.Subject, s.Email, s.Tokens, s.ExpiresAt = id.Subject, claims.Email, tok, time.Now().Add(time.Duration(tok.ExpiresIn)*time.Second)
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = c.store.Save(s); err != nil {
		return err
	}
	c.session = s
	return nil
}

func (c *Client) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session
	s.Tokens = Tokens{}
	s.ExpiresAt = time.Time{}
	if err := c.store.Save(s); err != nil {
		return err
	}
	c.session = s
	return nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session
	if !s.Connected() {
		return "", ErrPermission
	}
	if s.Tokens.Access != "" && time.Until(s.ExpiresAt) > time.Minute {
		return s.Tokens.Access, nil
	}
	tok, err := c.exchange(ctx, url.Values{"grant_type": {"refresh_token"}, "client_id": {s.ClientID}, "refresh_token": {s.Tokens.Refresh}, "resource": {resource}})
	if err != nil {
		return "", err
	}
	if tok.Refresh == "" {
		tok.Refresh = s.Tokens.Refresh
	}
	if tok.ID == "" {
		tok.ID = s.Tokens.ID
	}
	if tok.Scope == "" {
		tok.Scope = s.Tokens.Scope
	}
	s.Tokens, s.ExpiresAt = tok, time.Now().Add(time.Duration(tok.ExpiresIn)*time.Second)
	// Persist rotated credentials even if the new grant removed plan usage.
	if err = c.store.Save(s); err != nil {
		return "", err
	}
	c.session = s
	if !hasScope(tok.Scope, planScope) {
		return "", ErrPermission
	}
	return tok.Access, nil
}

func (c *Client) exchange(ctx context.Context, form url.Values) (Tokens, error) {
	var tok Tokens
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.issuer+"/api/accounts/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return tok, ErrAuth
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return tok, ctx.Err()
		}
		return tok, ErrNetwork
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return tok, ErrAuth
	}
	if json.NewDecoder(io.LimitReader(res.Body, 256<<10)).Decode(&tok) != nil || tok.Access == "" || !strings.EqualFold(tok.Type, "Bearer") || tok.ExpiresIn <= 0 || tok.ExpiresIn > 86400*30 {
		return Tokens{}, ErrAuth
	}
	return tok, nil
}

func hasScope(scopes, required string) bool {
	for _, s := range strings.Fields(scopes) {
		if s == required {
			return true
		}
	}
	return false
}
