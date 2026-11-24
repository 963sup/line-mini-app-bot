package openai

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	operations "github.com/963sup/line-mini-app-bot/internal/business/object/application/document"
	sdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func TestProcedureSuggestionsAreBoundedAndUntrusted(t *testing.T) {
	for _, tc := range []struct {
		name, status, text string
		want               error
	}{
		{"suggested", "completed", `{"candidates":[{"name":"基面處理","quote":"先確認基面","quantity":"99","unit":"SET","location":"invented"}]}`, nil},
		{"empty", "completed", `{"candidates":[]}`, operations.ErrInvalidOutput},
		{"invalid", "completed", `{"candidates":[{"name":"工序","quote":""}]}`, operations.ErrInvalidOutput},
		{"incomplete", "incomplete", "{", operations.ErrIncomplete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if body["store"] != false || body["tools"] != nil || body["max_output_tokens"] != float64(12000) {
					t.Error("unexpected model capabilities", body)
				}
				var input map[string]string
				if err := json.Unmarshal([]byte(body["input"].(string)), &input); err != nil || input["workItem"] != "防水" || input["conditions"] != "室內" {
					t.Error(input, err)
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]any{"id": "resp_test", "status": tc.status, "output": []any{map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]string{"type": "output_text", "text": tc.text}}}}}); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client := sdk.NewClient(option.WithAPIKey("test"), option.WithBaseURL(server.URL), option.WithMaxRetries(0))
			e := Extractor{Responses: client.Responses, Model: "test"}
			result, err := e.SuggestProcedures(t.Context(), "防水", "室內")
			if !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
			if err == nil && (len(result) != 1 || result[0].Quantity != nil || result[0].Unit != nil || result[0].Location != nil) {
				t.Fatal("model invented fields must not enter draft", result)
			}
		})
	}
}
