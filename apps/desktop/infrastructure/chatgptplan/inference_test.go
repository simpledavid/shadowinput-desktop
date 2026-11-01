package chatgptplan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arion.local/core/domain"
)

// Exercise the plan endpoint contract, including its rejection of the regular
// Responses API's string shorthand. See SIWC preview-limitations.
func TestGenerateUsesPlanMessageArrayAndCompletedStream(t *testing.T) {
	for _, contentType := range []string{"text/event-stream", ""} {
		t.Run("content-type="+contentType, func(t *testing.T) {
			const draft = "请把这句话写得自然一点。"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/responses" || r.Header.Get("Authorization") != "Bearer test-access" {
					t.Error("wrong route or authorization")
					http.Error(w, "bad route", http.StatusBadRequest)
					return
				}
				var body struct {
					Model        string                           `json:"model"`
					Input        []struct{ Role, Content string } `json:"input"`
					Instructions string                           `json:"instructions"`
					Store        *bool                            `json:"store"`
					Stream       bool                             `json:"stream"`
				}
				decoder := json.NewDecoder(r.Body)
				decoder.DisallowUnknownFields()
				if err := decoder.Decode(&body); err != nil {
					t.Errorf("unsupported plan request: %v", err)
					http.Error(w, `{"detail":"Input must be a list"}`, http.StatusBadRequest)
					return
				}
				if body.Model != "test-model" || len(body.Input) != 1 || body.Input[0].Role != "user" || body.Input[0].Content != draft || body.Instructions == "" || body.Store == nil || *body.Store || !body.Stream {
					t.Error("missing plan request requirements")
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if contentType == "" {
					w.Header()["Content-Type"] = nil // Suppress net/http automatic sniffing.
				} else {
					w.Header().Set("Content-Type", contentType)
				}
				_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"请让这句话更自然。\"}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\"}}\n\n"))
			}))
			defer server.Close()
			c, err := New(&memoryStore{session: Session{ClientID: "oaiapp_test", Subject: "test", Tokens: Tokens{Access: "test-access", Refresh: "test-refresh", Scope: planScope}, ExpiresAt: time.Now().Add(time.Hour)}})
			if err != nil {
				t.Fatal(err)
			}
			c.api, c.http = server.URL, server.Client()
			var preview string
			output, err := c.Generate(context.Background(), domain.WritingRequest{Task: "polish", Text: draft, Model: "test-model"}, func(text string) { preview = text })
			if err != nil || output != "请让这句话更自然。" || preview != output {
				t.Fatalf("plan generation did not complete: %q, %v", output, err)
			}
		})
	}
}
