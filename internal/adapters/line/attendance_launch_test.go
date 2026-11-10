package line

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestGroupCardOpensLIFFDirectly(t *testing.T) {
	s := &launchStub{attendanceStub: &attendanceStub{}}
	b := &Attendance{Store: s, AppURL: "https://liff.line.me/example"}
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
