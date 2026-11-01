package chatgptplan

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"arion.local/core/domain"
)

type Model struct {
	Slug       string `json:"slug"`
	Name       string `json:"display_name"`
	Visibility string `json:"visibility"`
}

func (c *Client) request(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.api+path, body)
	if err != nil {
		return nil, ErrNetwork
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrNetwork
	}
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		var failure struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(res.Body, 16384)).Decode(&failure)
		if failure.Error.Code == "model_not_found" || failure.Error.Code == "model_not_available" {
			return nil, ErrModel
		}
		if strings.HasPrefix(failure.Error.Code, "subscription_sharing_usage_") {
			return nil, ErrLimit
		}
		switch res.StatusCode {
		case 400, 422:
			return nil, ErrRequest
		case 401:
			return nil, ErrAuth
		case 403:
			return nil, ErrPermission
		case 429:
			return nil, ErrLimit
		}
		return nil, ErrNetwork
	}
	return res, nil
}

func (c *Client) Models(ctx context.Context) ([]Model, error) {
	res, err := c.request(ctx, http.MethodGet, "/models", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var catalog struct {
		Models []Model `json:"models"`
	}
	if json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(&catalog) != nil {
		return nil, ErrNetwork
	}
	out := make([]Model, 0, len(catalog.Models))
	for _, m := range catalog.Models {
		if m.Visibility == "list" && m.Slug != "" && m.Name != "" {
			out = append(out, m)
		}
	}
	if len(out) == 0 {
		return nil, ErrPermission
	}
	return out, nil
}

func (c *Client) Generate(ctx context.Context, r domain.WritingRequest, preview func(string)) (string, error) {
	if err := r.Validate(); err != nil {
		return "", err
	}
	// The ChatGPT plan route requires a message array, unlike the API-key
	// route's optional string shorthand. Keep these requirements explicit.
	body, err := json.Marshal(map[string]any{"model": r.Model, "input": []map[string]string{{"role": "user", "content": r.Text}}, "instructions": domain.WritingInstructions(r), "store": false, "stream": true})
	if err != nil {
		return "", err
	}
	res, err := c.request(ctx, http.MethodPost, "/responses", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "text/event-stream") {
		return "", ErrIncomplete
	}
	return readStream(res.Body, preview)
}

// readStream accepts output only at response.completed, never on EOF or [DONE].
func readStream(reader io.Reader, preview func(string)) (string, error) {
	scan := bufio.NewScanner(io.LimitReader(reader, 16<<20))
	scan.Buffer(make([]byte, 4096), 2<<20)
	var data []string
	var text strings.Builder
	consume := func() (bool, error) {
		if len(data) == 0 {
			return false, nil
		}
		raw := strings.Join(data, "\n")
		data = nil
		var event struct {
			Type     string `json:"type"`
			Delta    string `json:"delta"`
			Response struct {
				Status string `json:"status"`
				Error  struct {
					Code string `json:"code"`
				} `json:"error"`
			} `json:"response"`
		}
		if json.Unmarshal([]byte(raw), &event) != nil {
			return false, ErrIncomplete
		}
		switch event.Type {
		case "response.output_text.delta":
			if text.Len()+len(event.Delta) > 256000 {
				return false, ErrIncomplete
			}
			text.WriteString(event.Delta)
			if preview != nil {
				preview(text.String())
			}
		case "response.completed":
			if event.Response.Status != "completed" || strings.TrimSpace(text.String()) == "" {
				return false, ErrIncomplete
			}
			return true, nil
		case "response.failed":
			if strings.HasPrefix(event.Response.Error.Code, "subscription_sharing_usage_") {
				return false, ErrLimit
			}
			return false, ErrIncomplete
		case "response.incomplete", "error", "response.refusal.delta", "response.refusal.done":
			return false, ErrIncomplete
		}
		return false, nil
	}
	for scan.Scan() {
		line := scan.Text()
		if line == "" {
			done, err := consume()
			if err != nil {
				return "", err
			}
			if done {
				return text.String(), nil
			}
		} else if strings.HasPrefix(line, "data:") {
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	if scan.Err() != nil {
		return "", ErrIncomplete
	}
	return "", ErrIncomplete
}
