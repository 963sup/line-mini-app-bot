package summarizer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGeminiBoundsAndSources(t *testing.T) {
	const input = `{"Source":"S1","Speaker":"成員1","Text":"聊天"}`
	for _, tc := range []struct {
		name, body string
		ok         bool
	}{
		{"valid", `{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":"討論重點 [S1]"}]}}]}`, true},
		{"truncated", `{"candidates":[{"finishReason":"MAX_TOKENS","content":{"parts":[{"text":"摘要 [S1]"}]}}]}`, false},
		{"blocked", `{"candidates":[]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := Gemini{Key: "key", Model: "configured-model", HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				b, _ := io.ReadAll(r.Body)
				encoded, _ := json.Marshal(input)
				if r.Header.Get("x-goog-api-key") != "key" || !strings.Contains(string(b), string(encoded)) || !strings.Contains(string(b), "systemInstruction") {
					t.Fatal("missing prepared input or system instruction")
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}}
			_, err := g.Summarize(context.Background(), input)
			if (err == nil) != tc.ok {
				t.Fatalf("err=%v", err)
			}
		})
	}
}
