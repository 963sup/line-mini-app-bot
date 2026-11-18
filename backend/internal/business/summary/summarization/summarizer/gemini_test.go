package summarizer

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/summary/messagecollection"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGeminiBoundsAndSources(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		ok         bool
	}{
		{"valid", `{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":"討論重點 [S1]"}]}}]}`, true},
		{"invented citation", `{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":"來源 [S99]"}]}}]}`, false},
		{"truncated", `{"candidates":[{"finishReason":"MAX_TOKENS","content":{"parts":[{"text":"摘要 [S1]"}]}}]}`, false},
		{"blocked", `{"candidates":[]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := Gemini{Key: "key", Model: "configured-model", HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				b, _ := io.ReadAll(r.Body)
				if r.Header.Get("x-goog-api-key") != "key" || strings.Contains(string(b), "raw-user-id") || !strings.Contains(string(b), "systemInstruction") {
					t.Fatal("must separate instructions and alias raw IDs")
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}}
			_, err := g.Summarize(context.Background(), []messagecollection.Message{{ID: "source", UserID: "raw-user-id", Text: "聊天"}})
			if (err == nil) != tc.ok {
				t.Fatalf("err=%v", err)
			}
		})
	}
}
