package summary

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestOpenAISummaryContract(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		valid      bool
	}{
		{"valid", `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"討論重點 [S1]"}}]}`, 200, true},
		{"invented source", `{"choices":[{"finish_reason":"stop","message":{"content":"摘要 [S99]"}}]}`, 200, false},
		{"missing source", `{"choices":[{"finish_reason":"stop","message":{"content":"摘要"}}]}`, 200, false},
		{"truncated", `{"choices":[{"finish_reason":"length","message":{"content":"摘要 [S1]"}}]}`, 200, false},
		{"empty choices", `{"choices":[]}`, 200, false},
		{"refusal", `{"choices":[{"finish_reason":"stop","message":{"content":"摘要 [S1]","refusal":"refused"}}]}`, 200, false},
		{"malformed", `{`, 200, false},
		{"rate limited", `{"error":{"message":"private-provider-detail","type":"rate_limit"}}`, 429, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := OpenAI{Key: "test-key", Model: "gpt-4.1-mini", HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() != "https://api.openai.com/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
					t.Fatal("wrong endpoint or authentication")
				}
				deadline, ok := r.Context().Deadline()
				if !ok || time.Until(deadline) > 45*time.Second {
					t.Fatal("model call must have bounded deadline")
				}
				var body struct {
					Model               string
					MaxCompletionTokens int `json:"max_completion_tokens"`
					Messages            []struct{ Role, Content string }
					Tools               []any
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Model != gptModel || body.MaxCompletionTokens != 1500 || len(body.Messages) != 2 || body.Messages[0].Role != "system" || body.Messages[1].Role != "user" || len(body.Tools) != 0 {
					t.Fatal("invalid model request contract")
				}
				if strings.Contains(body.Messages[1].Content, "raw-user-id") || !strings.Contains(body.Messages[1].Content, "成員1") {
					t.Fatal("must alias participant IDs")
				}
				return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})}}
			_, err := g.Summarize(context.Background(), []Message{{ID: "source", UserID: "raw-user-id", Text: "聊天"}})
			if (err == nil) != tc.valid {
				t.Fatalf("unexpected result: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "private-provider-detail") {
				t.Fatal("provider error leaked")
			}
		})
	}
}

const gptModel = "gpt-4.1-mini"

func TestSummaryProviderSelection(t *testing.T) {
	for _, provider := range []string{"", "gemini", "openai"} {
		model, err := NewModel(provider, "key", "configured-model", http.DefaultClient)
		if err != nil {
			t.Fatal(err)
		}
		if provider == "openai" {
			if _, ok := model.(*OpenAI); !ok {
				t.Fatal("OpenAI not selected")
			}
		} else if _, ok := model.(*Gemini); !ok {
			t.Fatal("Gemini compatibility lost")
		}
	}
	for _, tc := range []struct{ provider, key, model string }{{"other", "key", "model"}, {"openai", "", "model"}, {"openai", "key", ""}} {
		if _, err := NewModel(tc.provider, tc.key, tc.model, http.DefaultClient); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}
