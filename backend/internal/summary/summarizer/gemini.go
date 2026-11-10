package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/summary"
)

type Gemini struct {
	Key, Model string
	HTTP       *http.Client
}

func (g *Gemini) Summarize(ctx context.Context, messages []summary.Message) (string, error) {
	input, sources := summary.Input(messages)
	body, _ := json.Marshal(map[string]any{
		"systemInstruction": map[string]any{"parts": []map[string]string{{"text": summaryInstruction}}},
		"contents":          []any{map[string]any{"role": "user", "parts": []map[string]string{{"text": input}}}},
		"generationConfig":  map[string]any{"maxOutputTokens": 1500},
	})
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://generativelanguage.googleapis.com/v1beta/models/"+url.PathEscape(g.Model)+":generateContent", bytes.NewReader(body))
	if e != nil {
		return "", errors.New("cannot build summary request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.Key)
	res, e := g.HTTP.Do(req)
	if e != nil {
		return "", errors.New("summary model unavailable")
	}
	// The response is read-only; closing cannot commit application state.
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return "", errors.New("summary model rejected request")
	}
	b, e := io.ReadAll(io.LimitReader(res.Body, (512<<10)+1))
	if e != nil || len(b) > 512<<10 {
		return "", errors.New("invalid model response size")
	}
	var output struct {
		Candidates []struct {
			FinishReason string
			Content      struct {
				Parts []struct {
					Text    string
					Thought bool
				}
			}
		}
	}
	if e = json.Unmarshal(b, &output); e != nil || len(output.Candidates) == 0 || output.Candidates[0].FinishReason != "STOP" {
		return "", errors.New("model did not finish summary")
	}
	var text strings.Builder
	for _, p := range output.Candidates[0].Content.Parts {
		if !p.Thought {
			text.WriteString(p.Text)
		}
	}
	return validateSummary(text.String(), len(sources))
}
