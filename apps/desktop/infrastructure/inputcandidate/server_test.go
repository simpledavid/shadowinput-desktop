package inputcandidate

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBoundaryAndCompletedCandidate(t *testing.T) {
	var calls atomic.Int32
	s, e := Start(t.TempDir(), func(ctx context.Context, text string) (string, error) {
		calls.Add(1)
		return "明天一起讨论一下", nil
	})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	b, _ := os.ReadFile(s.file)
	lines := strings.Split(string(b), "\n")
	for _, tc := range []struct {
		auth, origin, text string
		status             int
	}{{"", "", "明天讨论一下", 403}, {s.token, "https://evil.example", "明天讨论一下", 403}, {s.token, "", "hi", 400}, {s.token, "", "明天讨论一下", 200}} {
		req, _ := http.NewRequest("POST", "http://"+lines[0]+"/candidate", strings.NewReader(tc.text))
		if tc.auth != "" {
			req.Header.Set("Authorization", "Bearer "+tc.auth)
		}
		if tc.origin != "" {
			req.Header.Set("Origin", tc.origin)
		}
		r, e := http.DefaultClient.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		if r.StatusCode != tc.status {
			t.Fatalf("status %d want %d", r.StatusCode, tc.status)
		}
		if tc.status == 200 && string(body) != "明天一起讨论一下" {
			t.Fatal("missing result")
		}
	}
	if calls.Load() != 1 {
		t.Fatal("unauthorized request invoked model")
	}
}
func TestRevokeCancelsAndDropsLateResult(t *testing.T) {
	started := make(chan struct{})
	s, e := Start(t.TempDir(), func(ctx context.Context, text string) (string, error) {
		close(started)
		<-ctx.Done()
		return "不应显示的迟到结果", nil
	})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	b, _ := os.ReadFile(s.file)
	address := strings.Split(string(b), "\n")[0]
	done := make(chan int, 1)
	go func() {
		req, _ := http.NewRequest("POST", "http://"+address+"/candidate", strings.NewReader("这是组合短句"))
		req.Header.Set("Authorization", "Bearer "+s.token)
		r, e := http.DefaultClient.Do(req)
		if e != nil {
			done <- 0
			return
		}
		r.Body.Close()
		done <- r.StatusCode
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("not started")
	}
	s.Revoke()
	select {
	case status := <-done:
		if status != 204 {
			t.Fatalf("stale result status %d", status)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation blocked")
	}
}
func TestRejectInvalidCandidate(t *testing.T) {
	for _, text := range []string{"", "原来内容", "第一行\n第二行", "包含\t制表符", "包含\x1b控制符", strings.Repeat("字", 129), "<think>内部"} {
		if ValidResult("原来内容", text) {
			t.Fatalf("accepted %q", text)
		}
	}
}

func TestDisabledCandidateServiceRemovesDiscoveryAndRejectsSavedCapability(t *testing.T) {
	var calls atomic.Int32
	s, err := Start(t.TempDir(), func(context.Context, string) (string, error) { calls.Add(1); return "新的候选内容", nil })
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.file); !os.IsNotExist(err) {
		t.Fatal("discovery remains after opt-out")
	}
	req, _ := http.NewRequest("POST", "http://"+s.address+"/candidate", strings.NewReader("这是组合短句"))
	req.Header.Set("Authorization", "Bearer "+s.token)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	r.Body.Close()
	if r.StatusCode != http.StatusNoContent || calls.Load() != 0 {
		t.Fatal("opt-out invoked generation")
	}
}
