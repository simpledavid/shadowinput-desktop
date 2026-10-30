package application

import (
	"arion.local/core/domain"
	"context"
	"errors"
	"strings"
	"testing"
)

type writingProbe struct{ calls int }

func (p *writingProbe) Generate(context.Context, domain.WritingRequest, func(string)) (string, error) {
	p.calls++
	return "done", nil
}

func TestWritingRejectsInvalidInputBeforeSending(t *testing.T) {
	p := &writingProbe{}
	w := Writer{Generator: p}
	for _, r := range []domain.WritingRequest{{Task: "polish", Model: "m"}, {Task: "polish", Text: strings.Repeat("x", 64001), Model: "m"}, {Task: "other", Text: "text", Model: "m"}, {Task: "translate", Text: "text", Model: "m"}} {
		if _, err := w.Write(context.Background(), r, nil); !errors.Is(err, domain.ErrWritingInput) {
			t.Fatalf("invalid input accepted: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := w.Write(ctx, domain.WritingRequest{Task: "polish", Text: "text", Model: "m"}, nil)
	if !errors.Is(err, context.Canceled) || p.calls != 0 {
		t.Fatal("canceled or invalid input reached generator")
	}
}
