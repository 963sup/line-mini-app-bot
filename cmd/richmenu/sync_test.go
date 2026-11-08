package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type cleanupTransport func(*http.Request) (*http.Response, error)

func (f cleanupTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCleanupMigratesBeforeDelete(t *testing.T) {
	for _, phase := range []string{"succeeded", "failed"} {
		t.Run(phase, func(t *testing.T) {
			migrated, deleted := false, false
			transport := cleanupTransport(func(r *http.Request) (*http.Response, error) {
				body := `{}`
				switch {
				case strings.HasSuffix(r.URL.Path, "/richmenu/list"):
					body = `{"richmenus":[{"richMenuId":"old","name":"managed-attendance-in-` + strings.Repeat("a", 64) + `"},{"richMenuId":"foreign","name":"other-tool"},{"richMenuId":"new","name":"managed-attendance-in-` + strings.Repeat("b", 64) + `"}]}`
				case strings.HasSuffix(r.URL.Path, "/alias/list"):
					body = `{"aliases":[]}`
				case strings.HasSuffix(r.URL.Path, "/richmenu/batch"):
					data, _ := io.ReadAll(r.Body)
					if !strings.Contains(string(data), `"from":"old"`) || !strings.Contains(string(data), `"to":"new"`) {
						t.Fatal("wrong migration")
					}
					migrated = true
				case strings.HasSuffix(r.URL.Path, "/progress/batch"):
					body = `{"phase":"` + phase + `"}`
				case r.Method == "DELETE" && strings.HasSuffix(r.URL.Path, "/richmenu/old"):
					if !migrated || phase != "succeeded" {
						t.Fatal("deleted before successful migration")
					}
					deleted = true
				default:
					t.Fatalf("unexpected operation %s %s", r.Method, r.URL.Path)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"X-Line-Request-Id": []string{"request"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			api, err := messaging_api.NewMessagingApiAPI("test", messaging_api.WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatal(err)
			}
			err = cleanupMenus(api, map[string]string{"attendance-in": "new", "attendance-out": "new-out"})
			if phase == "succeeded" && (err != nil || !deleted) {
				t.Fatalf("cleanup failed: %v", err)
			}
			if phase == "failed" && (err == nil || deleted) {
				t.Fatal("failed migration must retain old menu")
			}
		})
	}
}
