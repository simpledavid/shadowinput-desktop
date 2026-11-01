package localwriting

import (
	"context"
	"errors"
	"strings"
	"testing"

	"arion.local/core/domain"
	"arion.local/desktop/infrastructure/llm/engine"
	"arion.local/desktop/infrastructure/llm/prompt"
	"arion.local/desktop/infrastructure/rewriting"
)

type completeFunc func(context.Context, prompt.ChatRequest) (engine.Result, error)

func (f completeFunc) Complete(c context.Context, r prompt.ChatRequest) (engine.Result, error) {
	return f(c, r)
}
func request() domain.WritingRequest {
	return domain.WritingRequest{Task: "polish", Text: "Please check shadowinput by Friday.", Model: rewriting.DefaultModel}
}

func TestLocalWritingSharesTaskAndSelectedPreferences(t *testing.T) {
	for _, task := range []string{"polish", "translate", "rewrite", "continue"} {
		r := request()
		r.Task = task
		r.Language = "Chinese"
		r.Terms = []string{"shadowinput", "unrelated secret"}
		style, _ := domain.DefaultWritingProfile().Style("work")
		style.Rules = "Use short sentences."
		r.Style = &style
		calls, previews := 0, 0
		g := Generator{Chat: completeFunc(func(_ context.Context, p prompt.ChatRequest) (engine.Result, error) {
			calls++
			if len(p.Messages) != 2 || p.Messages[0].Role != "system" || p.Messages[0].Content != domain.WritingInstructions(r) || p.Messages[1].Content != r.Text || strings.Contains(p.Messages[0].Content, "unrelated secret") || p.AllowThinking || p.Stream || p.MaxTokens < 512 || p.MaxTokens > 2048 {
				t.Fatalf("bad request: %+v", p)
			}
			return engine.Result{Text: " Complete result "}, nil
		})}
		got, err := g.Generate(context.Background(), r, func(text string) {
			previews++
			if text != "Complete result" {
				t.Fatal(text)
			}
		})
		if err != nil || got != "Complete result" || calls != 1 || previews != 1 {
			t.Fatalf("%s: %q %v %d/%d", task, got, err, calls, previews)
		}
	}
}

func TestIncompleteAndFailedLocalResultsNeverPreview(t *testing.T) {
	for _, tc := range []struct {
		result engine.Result
		err    error
	}{
		{result: engine.Result{Text: "partial", Truncated: true}}, {result: engine.Result{Text: "partial", Cancelled: true}},
		{result: engine.Result{Text: ""}}, {result: engine.Result{Text: "<think>hidden</think>result"}},
		{result: engine.Result{Text: "bad<|im_end|>"}}, {result: engine.Result{Text: "bad\x00"}},
		{result: engine.Result{Text: "partial"}, err: engine.ErrOutputBudget},
		{err: engine.ErrInputBudget},
	} {
		g := Generator{Chat: completeFunc(func(context.Context, prompt.ChatRequest) (engine.Result, error) { return tc.result, tc.err })}
		got, err := g.Generate(context.Background(), request(), func(string) { t.Fatal("previewed failure") })
		if err == nil || got != "" {
			t.Fatalf("accepted %+v: %q %v", tc, got, err)
		}
		if tc.err != nil && !errors.Is(err, tc.err) {
			t.Fatal("lost typed error")
		}
	}
}

func TestInvalidAndCancelledRequestsNeverReachWorker(t *testing.T) {
	g := Generator{Chat: completeFunc(func(context.Context, prompt.ChatRequest) (engine.Result, error) {
		t.Fatal("worker called")
		return engine.Result{}, nil
	})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := g.Generate(ctx, request(), nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, field := range []string{"source", "instruction", "model"} {
		r := request()
		switch field {
		case "source":
			r.Text = "bad<|im_start|>"
		case "instruction":
			r.Instruction = "bad\x00"
		case "model":
			r.Model = "cloud-model"
		}
		if _, err := g.Generate(context.Background(), r, nil); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
}

func TestPhraseCompletionHasShortBudgetAndSuffixContract(t *testing.T) {
	r := request()
	r.Task = "complete_phrase"
	r.Text = "麻烦帮我"
	g := Generator{Chat: completeFunc(func(_ context.Context, p prompt.ChatRequest) (engine.Result, error) {
		if p.MaxTokens > 64 || p.AllowThinking || p.Stream || !strings.Contains(p.Messages[0].Content, "only the missing suffix") {
			t.Fatalf("not a fast suffix request: %+v", p)
		}
		return engine.Result{Text: "确认一下"}, nil
	})}
	text, err := g.Generate(context.Background(), r, nil)
	if err != nil || text != "确认一下" {
		t.Fatalf("%q %v", text, err)
	}
}
