// Package localwriting adapts the pinned offline chat worker to writing tasks.
package localwriting

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"arion.local/core/domain"
	"arion.local/core/domain/ports"
	"arion.local/desktop/infrastructure/llm/engine"
	"arion.local/desktop/infrastructure/llm/prompt"
	"arion.local/desktop/infrastructure/rewriting"
)

var ErrIncomplete = errors.New("local writing did not complete")

type Completer interface {
	Complete(context.Context, prompt.ChatRequest) (engine.Result, error)
}

// Generator uses an already enabled worker. It never downloads or chooses cloud.
type Generator struct{ Chat Completer }

func Prepare(r domain.WritingRequest) (prompt.ChatRequest, error) {
	if err := r.Validate(); err != nil {
		return prompt.ChatRequest{}, err
	}
	if r.Model != rewriting.DefaultModel {
		return prompt.ChatRequest{}, domain.ErrWritingInput
	}
	instructions := domain.WritingInstructions(r)
	for _, text := range []string{r.Text, instructions} {
		if !utf8.ValidString(text) || strings.ContainsAny(text, "\x00") || strings.Contains(text, "<|") || strings.Contains(text, "|>") {
			return prompt.ChatRequest{}, engine.ErrControlInput
		}
	}
	budget := min(2048, max(512, utf8.RuneCountInString(r.Text)*2+256))
	if r.Task == domain.TaskCompletePhrase {
		budget = 64
	}
	return prompt.ChatRequest{
		Messages: []prompt.Message{{Role: "system", Content: instructions}, {Role: "user", Content: r.Text}},
		Seed:     42, MaxTokens: budget,
	}, nil
}

func (g Generator) Generate(ctx context.Context, r domain.WritingRequest, preview func(string)) (string, error) {
	req, err := Prepare(r)
	if err != nil {
		return "", err
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	if g.Chat == nil {
		return "", domain.Fail(domain.CodeModelMissing, "local writing unavailable")
	}
	result, err := g.Chat.Complete(ctx, req)
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		return "", err
	}
	if result.Cancelled {
		return "", context.Canceled
	}
	text := strings.TrimSpace(result.Text)
	if result.Truncated || text == "" || len(text) > 256000 || !utf8.ValidString(text) || strings.ContainsAny(text, "\x00") ||
		strings.Contains(text, "<|") || strings.Contains(text, "|>") || strings.Contains(strings.ToLower(text), "<think") || strings.Contains(strings.ToLower(text), "</think") {
		return "", ErrIncomplete
	}
	// The worker exposes complete responses only; incomplete text never previews.
	if preview != nil {
		preview(text)
	}
	return text, nil
}

var _ ports.WritingGenerator = Generator{}
