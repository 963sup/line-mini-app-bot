package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
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

type launchStub struct {
	*attendanceStub
	actions []string
}

func (s *launchStub) CreateLaunch(_ context.Context, _, action string) (string, error) {
	s.actions = append(s.actions, action)
	return strings.Repeat(string('a'+byte(len(s.actions)-1)), 64), nil
}
func (s *launchStub) LaunchGroup(context.Context, string) (string, error) { return "group", nil }
func (s *launchStub) ExchangeLaunch(context.Context, string, string) (string, error) {
	return strings.Repeat("f", 64), nil
}

func TestGroupCardOpensLIFFDirectly(t *testing.T) {
	s := &launchStub{attendanceStub: &attendanceStub{}}
	b := &attendanceBot{store: s, appURL: "https://liff.line.me/example"}
	e := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", WebhookEventId: "event", Source: webhook.GroupSource{GroupId: "group", UserId: "actor"}, Message: webhook.TextMessageContent{Text: "打卡"}}
	reply, handled, err := b.reply(context.Background(), e)
	if err != nil || !handled || len(s.actions) != 5 || s.actions[0] != "clock-in" || s.actions[1] != "clock-out" || s.actions[4] != "confirm-rest" {
		t.Fatal("card must mint both actions")
	}
	data, err := json.Marshal(reply.message)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "https://liff.line.me/example?launch=") || strings.Contains(string(data), "action=clock-in") || s.calls != 0 {
		t.Fatalf("card must open LIFF directly without punching: %s", data)
	}
}
