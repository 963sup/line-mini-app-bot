package openai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/objectimport/parsing"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

const candidate = `{"rows":[{"name":"Line","description":"proposal","type":"manual","quantity":"11.00","unit":"SET","materialCode":"","sourceItemNo":"10","evidence":[{"page":2,"quote":"literal"}]}]}`

func response(content, finish, refusal string, tools []any) *http.Response {
	body, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": finish, "message": map[string]any{"role": "assistant", "content": content, "refusal": refusal, "tool_calls": tools}}}})
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(string(body))), ContentLength: int64(len(body))}
}

func TestProviderStrictRequestAndRejectedOutput(t *testing.T) {
	cases := []struct {
		name, content, finish, refusal string
		tools                          []any
		ok                             bool
	}{{"valid", candidate, "stop", "", nil, true}, {"refusal", candidate, "stop", "declined", nil, false}, {"truncated", candidate, "length", "", nil, false}, {"tools", candidate, "stop", "", []any{map[string]any{"id": "call", "type": "function", "function": map[string]any{"name": "not_allowed", "arguments": "{}"}}}, false}, {"malformed", "{", "stop", "", nil, false}, {"unknown authority", strings.Replace(candidate, `"name":"Line"`, `"status":"completed","name":"Line"`, 1), "stop", "", nil, false}, {"missing required", strings.Replace(candidate, `"materialCode":"",`, "", 1), "stop", "", nil, false}, {"oversize", strings.Repeat("x", MaxOutputBytes+1), "stop", "", nil, false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			p := Producer{Key: "fake-key", HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != "api.openai.com" || r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer fake-key" {
					t.Fatal("wrong provider request")
				}
				var input map[string]any
				if e := json.NewDecoder(r.Body).Decode(&input); e != nil {
					t.Fatal(e)
				}
				format := input["response_format"].(map[string]any)
				strict := format["json_schema"].(map[string]any)
				if input["model"] != DefaultModel || input["max_completion_tokens"] != float64(16000) || format["type"] != "json_schema" || strict["strict"] != true {
					t.Fatal("unbounded/unstructured request", input)
				}
				return response(tc.content, tc.finish, tc.refusal, tc.tools), nil
			})}}
			rows, e := p.Candidates(context.Background(), feature.Source{Pages: []feature.Page{{Number: 2, Text: "literal source"}}})
			if tc.ok {
				if e != nil || len(rows) != 1 || rows[0].Quantity != "11.00" {
					t.Fatal(rows, e)
				}
			} else if !errors.Is(e, parsing.ErrOutput) {
				t.Fatal("unsafe output accepted", e)
			}
			if calls != 1 {
				t.Fatal(calls)
			}
		})
	}
}

func TestProviderCancellationBoundsAndConfiguration(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	p := Producer{HTTP: client}
	if _, e := p.Candidates(context.Background(), feature.Source{}); !errors.Is(e, parsing.ErrNotConfigured) || calls != 0 {
		t.Fatal("missing credentials made request", e)
	}
	p.Key = "fake-key"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := p.Candidates(ctx, feature.Source{}); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, e := p.Candidates(ctx, feature.Source{}); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	p.HTTP = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		r := response(candidate, "stop", "", nil)
		r.ContentLength = maxResponseBytes + 1
		return r, nil
	})}
	if _, e := p.Candidates(context.Background(), feature.Source{}); !errors.Is(e, parsing.ErrModel) {
		t.Fatal("unbounded response", e)
	}
}
