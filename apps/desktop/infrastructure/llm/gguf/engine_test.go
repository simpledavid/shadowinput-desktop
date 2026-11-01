package gguf

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"arion.local/desktop/infrastructure/llm/engine"
	"arion.local/desktop/infrastructure/llm/prompt"
)

func TestFailedNativeReloadCannotContinueServing(t *testing.T) {
	e := &Engine{reload: true, runtimeDir: t.TempDir(), config: engine.Config{ModelDir: t.TempDir(), Threads: 1, MaxLength: 4096, MaxNewTokens: 256}}
	_, err := e.Run(engine.Request{Session: "after-cancellation"}, &atomic.Bool{})
	if !errors.Is(err, ErrRuntimeUnavailable) {
		t.Fatalf("failed native reload was not fatal: %v", err)
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPreparationNeverSamplesToken(t *testing.T) {
	calls := []string{}
	e := &Engine{config: engine.Config{MaxLength: 4096}, http: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, r.URL.Path)
		body := `{"prompt":"fully rendered native prompt"}`
		if r.URL.Path == "/tokenize" {
			body = `{"tokens":[1,2,3]}`
		}
		if r.URL.Path == "/v1/chat/completions" {
			t.Fatal("prefill attempted token generation")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})}}
	result, err := e.generate(context.Background(), prompt.ChatRequest{MaxTokens: 256}, true)
	if err != nil || result.Usage.Input != 3 || result.Usage.Output != 0 || len(calls) != 2 {
		t.Fatalf("result=%+v calls=%v error=%v", result, calls, err)
	}
}

func TestOversizedContextFailsWithoutGeneration(t *testing.T) {
	e := &Engine{config: engine.Config{MaxLength: 258}, http: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		var body string
		switch r.URL.Path {
		case "/apply-template":
			body = `{"prompt":"full context"}`
		case "/tokenize":
			body = `{"tokens":[1,2,3]}`
		default:
			t.Fatal("oversized input reached generation")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	_, err := e.generate(context.Background(), prompt.ChatRequest{MaxTokens: 256}, false)
	if err != engine.ErrInputBudget {
		t.Fatal(err)
	}
}

func TestUnexpectedReasoningIsProtocolFailure(t *testing.T) {
	e := &Engine{config: engine.Config{MaxLength: 4096}, http: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		var body any
		switch r.URL.Path {
		case "/apply-template":
			body = map[string]any{"prompt": "p"}
		case "/tokenize":
			body = map[string]any{"tokens": []int{1}}
		default:
			body = map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "text", "reasoning_content": "private reasoning"}, "finish_reason": "stop"}}}
		}
		data, _ := json.Marshal(body)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(data)))}, nil
	})}}
	if _, err := e.generate(context.Background(), prompt.ChatRequest{MaxTokens: 256}, false); err == nil {
		t.Fatal("accepted unexpected reasoning")
	}
}

func TestConfiguredOutputBudgetIsEnforced(t *testing.T) {
	e := &Engine{config: engine.Config{MaxLength: 4096, MaxNewTokens: 8}, http: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		body := `{"prompt":"p"}`
		switch r.URL.Path {
		case "/tokenize":
			body = `{"tokens":[1]}`
		case "/v1/chat/completions":
			var request prompt.ChatRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			enabled, ok := request.TemplateArgs["enable_thinking"]
			if request.MaxTokens != 8 || !ok || enabled {
				t.Fatalf("token cap %d thinking %v ok %v", request.MaxTokens, enabled, ok)
			}
			body = `{"choices":[{"message":{"content":"text"},"finish_reason":"stop"}]}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	if _, err := e.generate(context.Background(), prompt.ChatRequest{MaxTokens: 256}, false); err != nil {
		t.Fatal(err)
	}
}

func TestIndependentCompletionErasesBothSidesAndAllowsWritingBudget(t *testing.T) {
	for _, finish := range []string{"stop", "length"} {
		calls := []string{}
		e := &Engine{config: engine.Config{MaxLength: 4096, MaxNewTokens: 2048}, http: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			calls = append(calls, r.URL.Path)
			body := `{}`
			switch r.URL.Path {
			case "/apply-template":
				body = `{"prompt":"p"}`
			case "/tokenize":
				body = `{"tokens":[1,2,3]}`
			case "/v1/chat/completions":
				var req prompt.ChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatal(err)
				}
				if req.MaxTokens != 1024 || req.CachePrompt || req.Stream || req.TemplateArgs["enable_thinking"] {
					t.Fatalf("unsafe request %+v", req)
				}
				body = `{"choices":[{"message":{"content":"complete"},"finish_reason":"` + finish + `"}]}`
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
		})}}
		result, err := e.Complete(context.Background(), prompt.ChatRequest{Messages: []prompt.Message{{Role: "user", Content: "draft"}}, MaxTokens: 1024, AllowThinking: true, CachePrompt: true, Stream: true})
		if finish == "stop" && (err != nil || result.Text != "complete") {
			t.Fatal(result, err)
		}
		if finish == "length" && (!errors.Is(err, engine.ErrOutputBudget) || result.Text != "") {
			t.Fatal(result, err)
		}
		if len(calls) != 5 || calls[0] != "/slots/0" || calls[4] != "/slots/0" {
			t.Fatalf("state was not erased: %v", calls)
		}
	}
}
