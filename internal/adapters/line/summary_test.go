package line

import (
	"context"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/summary"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type chatStub struct {
	summary.Store
	collected           []summary.Message
	group               string
	commands, withdrawn int
	job                 summary.Job
}

func (s *chatStub) Collect(_ context.Context, g string, m summary.Message) error {
	s.group = g
	s.collected = append(s.collected, m)
	return nil
}

func (s *chatStub) Command(_ context.Context, g, _, _ string, _ int64) (summary.Job, error) {
	s.group = g
	s.commands++
	return s.job, nil
}
func (s *chatStub) Unsend(context.Context, string, string) error { s.withdrawn++; return nil }

type queueStub struct{ calls int }

func (q *queueStub) Enqueue(context.Context, summary.Job) error { q.calls++; return nil }

func TestSummaryCollectsGroupTextSilentlyWithActor(t *testing.T) {
	s := &chatStub{}
	b := &Summary{Store: s}
	e := webhook.MessageEvent{Mode: "active", Timestamp: time.Now().UnixMilli(), Source: webhook.GroupSource{GroupId: "g", UserId: "u"}, Message: webhook.TextMessageContent{Id: "m", Text: "今天討論出貨"}, WebhookEventId: "event"}
	r, handled, err := b.observe(context.Background(), e)
	if err != nil || handled || r.message != nil || len(s.collected) != 1 || s.collected[0].UserID != "u" || s.collected[0].ID != "m" || s.group != "g" {
		t.Fatal("must collect silently using event identity")
	}
	e.Source = webhook.UserSource{UserId: "u"}
	_, _, _ = b.observe(context.Background(), e)
	e.Source = webhook.GroupSource{GroupId: "g", UserId: "u"}
	e.Message = webhook.TextMessageContent{Text: "上班"}
	_, _, _ = b.observe(context.Background(), e)
	if len(s.collected) != 1 {
		t.Fatal("direct messages and attendance commands must not be collected")
	}
}

func TestSummaryEnqueuesWithoutCallingModel(t *testing.T) {
	s := &chatStub{job: summary.Job{Status: "pending"}}
	q := &queueStub{}
	b := &Summary{Store: s, Queue: q}
	e := webhook.MessageEvent{Mode: "active", Timestamp: time.Now().UnixMilli(), ReplyToken: "reply", WebhookEventId: "event", Source: webhook.GroupSource{GroupId: "g", UserId: "u"}, Message: webhook.TextMessageContent{Text: "摘要"}, DeliveryContext: &webhook.DeliveryContext{IsRedelivery: true}}
	r, handled, err := b.observe(context.Background(), e)
	if err != nil || !handled || r.message == nil || q.calls != 1 || s.commands != 1 {
		t.Fatal("must persist and enqueue a command, including redelivery")
	}
	s.job.Status = "done"
	_, _, _ = b.observe(context.Background(), e)
	if q.calls != 1 {
		t.Fatal("completed job must not be requeued")
	}
}

func TestLegacySummarySwitchCommandsDoNothing(t *testing.T) {
	s := &chatStub{}
	b := &Summary{Store: s}
	for _, command := range []string{"關閉摘要", "開啟摘要"} {
		e := webhook.MessageEvent{Mode: "active", Timestamp: time.Now().UnixMilli(), ReplyToken: "reply", WebhookEventId: "event", Source: webhook.GroupSource{GroupId: "g", UserId: "u"}, Message: webhook.TextMessageContent{Id: "m", Text: command}}
		reply, handled, err := b.observe(context.Background(), e)
		if err != nil || handled || reply.message != nil || s.commands != 0 || len(s.collected) != 0 {
			t.Fatalf("legacy command %q must be inert", command)
		}
	}
}
