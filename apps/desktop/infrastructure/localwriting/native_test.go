package localwriting

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"arion.local/core/domain"
	"arion.local/desktop/infrastructure/llm/engine"
	"arion.local/desktop/infrastructure/provision"
	"arion.local/desktop/infrastructure/rewriting"
	"arion.local/models"
)

// Opt-in smoke test of the actual packaged worker. This is not a quality benchmark.
func TestNativeWriting(t *testing.T) {
	dir, model := os.Getenv("SHADOWINPUT_LLM_RUNTIME"), os.Getenv("SHADOWINPUT_LLM_MODEL")
	if dir == "" || model == "" {
		t.Skip("pinned runtime and verified model required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	catalog, err := models.Load()
	if err != nil {
		t.Fatal(err)
	}
	local := rewriting.New(catalog, provision.NewStore(filepath.Dir(model)), dir, 4, runtime.GOOS+"-"+runtime.GOARCH)
	defer local.Disable()
	local.Refresh()
	if !local.Status().Installed {
		t.Fatal("verified model not found")
	}
	if err := local.Enable(ctx); err != nil {
		t.Fatal(err)
	}
	if !local.Ready() {
		t.Fatal("explicit offline start did not become ready")
	}
	g := Generator{Chat: local}

	style, _ := domain.DefaultWritingProfile().Style("work")
	for _, task := range []string{"polish", "translate", "rewrite", "continue"} {
		r := domain.WritingRequest{Task: task, Text: "请帮我确认 shadowinput 的测试安排，预计周五完成。", Language: "English", Model: rewriting.DefaultModel, Style: &style, Terms: []string{"shadowinput"}, Instruction: "Keep it brief; do not add dates or commitments."}
		started := time.Now()
		text, err := g.Generate(ctx, r, nil)
		t.Logf("%s elapsed=%s output=%q error=%v", task, time.Since(started).Round(time.Millisecond), text, err)
		if err != nil || text == "" || !strings.Contains(text, "shadowinput") {
			t.Fatalf("%s failed", task)
		}
	}
	r := request()
	r.Text = strings.Repeat("这是需要检查的文本。", 1500)
	if _, err := g.Generate(ctx, r, nil); !errors.Is(err, engine.ErrInputBudget) {
		t.Fatalf("context overflow lost: %v", err)
	}
	r = request()
	r.Task = "continue"
	r.Text = "The project plan has three stages. "
	r.Instruction = "Write a long continuation."
	short, stop := context.WithTimeout(ctx, 100*time.Millisecond)
	started := time.Now()
	_, err = g.Generate(short, r, nil)
	stop()
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 3*time.Second {
		t.Fatalf("cancel not bounded: %v", err)
	}
	text, err := g.Generate(ctx, request(), nil)
	if err != nil || text == "" {
		t.Fatalf("fresh request after cancellation failed: %q %v", text, err)
	}
	t.Logf("cancel and retry output=%q", text)
}
