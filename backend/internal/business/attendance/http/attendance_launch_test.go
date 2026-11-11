package attendancehttp

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type memberTransport func(*http.Request) (*http.Response, error)

func (f memberTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGroupMemberCheckFailsClosed(t *testing.T) {
	for _, code := range []int{200, 400, 404, 429, 500} {
		client := &http.Client{Transport: memberTransport(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v2/bot/group/group/member/user" || r.Header.Get("Authorization") != "Bearer test" {
				t.Error("incorrect membership request")
			}
			return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
		})}
		if got := checkGroupMember(context.Background(), client, "https://example.test", "test", "group", "user"); got != code {
			t.Fatalf("status %d got %d", code, got)
		}
	}
	client := &http.Client{Transport: memberTransport(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if got := checkGroupMember(ctx, client, "https://example.test", "test", "group", "user"); got != 0 {
		t.Fatalf("timeout must not authorize: %d", got)
	}
}
