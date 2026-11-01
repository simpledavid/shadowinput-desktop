// Package inputcandidate serves bounded, opt-in AI requests from native input methods.
// No keystrokes, document context, candidates or credentials are logged or persisted.
package inputcandidate

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

type Generate func(context.Context, string) (string, error)

type Server struct {
	enabled     bool
	address     string
	mu          sync.Mutex
	cancel      context.CancelFunc
	epoch       uint64
	token, file string
	http        *http.Server
	generate    Generate
}

// Start binds only an OS-assigned loopback port. The per-launch capability is
// stored with user-only permissions; browsers are refused even with a token.
func Start(directory string, generate Generate) (*Server, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &Server{enabled: true, address: listener.Addr().String(), token: rand.Text(), file: filepath.Join(directory, "candidate.endpoint"), generate: generate}
	s.http = &http.Server{Handler: http.HandlerFunc(s.serve), ReadHeaderTimeout: time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: time.Second, MaxHeaderBytes: 4096}
	if err = os.MkdirAll(directory, 0700); err == nil {
		err = os.WriteFile(s.file, []byte(listener.Addr().String()+"\n"+s.token+"\n"), 0600)
	}
	if err != nil {
		listener.Close()
		return nil, err
	}
	if err = os.Chmod(s.file, 0600); err != nil {
		listener.Close()
		os.Remove(s.file)
		return nil, err
	}
	go s.http.Serve(listener)
	return s, nil
}

// SetEnabled removes discovery when disabled and cancels in-flight inference.
func (s *Server) SetEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.enabled == enabled {
		return nil
	}
	s.epoch++
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if !enabled {
		s.enabled = false
		return os.Remove(s.file)
	}
	if err := os.WriteFile(s.file, []byte(s.address+"\n"+s.token+"\n"), 0600); err != nil {
		return err
	}
	s.enabled = true
	return nil
}

func (s *Server) Revoke() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.epoch++
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}
func (s *Server) Close() { s.Revoke(); s.http.Close(); os.Remove(s.file) }
func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != "POST" || r.URL.Path != "/candidate" || r.Header.Get("Origin") != "" || r.Header.Get("Sec-Fetch-Site") != "" ||
		subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+s.token)) != 1 {
		http.Error(w, "unavailable", http.StatusForbidden)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1025))
	text := strings.TrimSpace(string(body))
	if err != nil || len(body) > 1024 || !utf8.Valid(body) || strings.IndexFunc(text, unicode.IsControl) >= 0 || utf8.RuneCountInString(text) < 2 || utf8.RuneCountInString(text) > 128 {
		http.Error(w, "invalid continuation context", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	s.mu.Lock()
	if !s.enabled {
		s.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.epoch++
	epoch := s.epoch
	s.cancel = cancel
	s.mu.Unlock()
	result, err := s.generate(ctx, text)
	s.mu.Lock()
	current := s.epoch == epoch
	if current {
		s.cancel = nil
	}
	s.mu.Unlock()
	result = strings.TrimSpace(result)
	if err != nil || ctx.Err() != nil || !current || !ValidResult(text, result) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	io.WriteString(w, result)
}

func ValidResult(source, result string) bool {
	return result != "" && !strings.HasPrefix(result, source) && !strings.HasSuffix(source, result) && !repeatedBoundary(source, result) && !assistantReply(result) && utf8.ValidString(result) && utf8.RuneCountInString(result) <= 24 &&
		strings.IndexFunc(result, unicode.IsControl) < 0 && !strings.Contains(result, "<|") && !strings.Contains(result, "<think")
}

// Conservative surface checks, not a semantic quality guarantee.
func repeatedBoundary(source, result string) bool {
	prefix, suffix := []rune(source), []rune(result)
	for count := 2; count <= min(len(prefix), len(suffix)); count++ {
		if string(prefix[len(prefix)-count:]) == string(suffix[:count]) {
			return true
		}
	}
	return false
}
func assistantReply(text string) bool {
	for _, phrase := range []string{"请提供", "有什么可以帮", "需要我帮", "作为AI", "作为 AI"} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

var ErrUnavailable = errors.New("AI candidates unavailable")
