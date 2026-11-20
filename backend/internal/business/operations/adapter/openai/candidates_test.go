package openai

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	sdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func TestFileResponseContract(t *testing.T) {
	for _, tc := range []struct {
		name, status, kind, value string
		want                      error
		count                     int
	}{
		{"candidate", "completed", "output_text", `{"candidates":[{"name":"電纜","quantity":"2.50","unit":"M","location":null,"quote":"電纜 2.50 M"}]}`, nil, 1},
		{"empty", "completed", "output_text", `{"candidates":[]}`, nil, 0},
		{"refused", "completed", "refusal", "refused", operations.ErrRefused, 0},
		{"truncated", "incomplete", "output_text", "{", operations.ErrIncomplete, 0},
		{"invalid", "completed", "output_text", "{}", operations.ErrInvalidOutput, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Input []struct {
						Content []struct {
							Type, Filename string
							FileData       string `json:"file_data"`
						}
					}
					Text struct {
						Format struct {
							Type   string
							Strict bool
							Schema map[string]any
						}
					}
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if r.Method != "POST" || r.URL.Path != "/responses" || len(body.Input) != 1 || len(body.Input[0].Content) != 1 {
					t.Error("unexpected file request")
					return
				}
				file := body.Input[0].Content[0]
				if file.Type != "input_file" || file.Filename != "工項.txt" || file.FileData != "data:text/plain;base64,YWJj" || body.Text.Format.Type != "json_schema" || !body.Text.Format.Strict || body.Text.Format.Schema["additionalProperties"] != false {
					t.Error("unexpected file or output format")
				}
				part := map[string]string{"type": tc.kind, "text": tc.value}
				if tc.kind == "refusal" {
					part = map[string]string{"type": "refusal", "refusal": tc.value}
				}
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(map[string]any{"id": "resp_test", "status": tc.status, "output": []any{map[string]any{"type": "message", "role": "assistant", "content": []any{part}}}})
				if err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client := sdk.NewClient(option.WithAPIKey("test"), option.WithBaseURL(server.URL), option.WithMaxRetries(0))
			extractor := Extractor{Responses: client.Responses, Model: "test"}
			result, err := extractor.ExtractCandidates(t.Context(), operations.Document{Filename: "工項.txt", Data: []byte("abc")})
			if !errors.Is(err, tc.want) || len(result.Candidates) != tc.count || !json.Valid(result.RawResponse) {
				t.Fatalf("candidates=%v error=%v", result.Candidates, err)
			}
		})
	}
}
