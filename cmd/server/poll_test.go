package main

import (
	"context"
	"github.com/963sup/line-mini-app-bot/internal/poll"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"testing"
)

type pollStub struct {
	poll.Store
	calls int
}

func (s *pollStub) Vote(_ context.Context, group, user, event, id string, option int) (poll.Result, error) {
	s.calls++
	return poll.Result{Poll: poll.Poll{ID: id, Question: "午餐", Options: []string{"飯", "麵"}, Counts: []int{1, 0}}}, nil
}
func TestPollIdentityAndRedelivery(t *testing.T) {
	for _, user := range []string{"", "actor"} {
		s := &pollStub{}
		b := &pollBot{store: s}
		e := webhook.PostbackEvent{Mode: "active", WebhookEventId: "event", ReplyToken: "token", Source: webhook.GroupSource{GroupId: "group", UserId: user}, Postback: &webhook.PostbackContent{Data: "action=poll-vote&poll=poll1&option=0"}, DeliveryContext: &webhook.DeliveryContext{IsRedelivery: true}}
		_, handled, err := b.handle(context.Background(), e)
		want := 0
		if user != "" {
			want = 1
		}
		if err != nil || !handled || s.calls != want {
			t.Fatalf("user=%q calls=%d handled=%v err=%v", user, s.calls, handled, err)
		}
	}
}
