package line

import (
	"context"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/summary/messagecollection"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Collection struct {
	Store messagecollection.Store
}

// Observe is independent of command replies: ordinary text remains silent.
func (b *Collection) Observe(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
	if b == nil {
		return transport.Reply{}, false, nil
	}
	if e, ok := event.(webhook.MessageEditedEvent); ok {
		g, group := e.Source.(webhook.GroupSource)
		m, text := e.Message.(webhook.TextMessageContent)
		if editor, supported := b.Store.(interface {
			Edit(context.Context, string, string, string, int64) error
		}); supported && group && text && g.GroupId != "" && e.Mode == "active" {
			return transport.Reply{}, false, editor.Edit(ctx, g.GroupId, m.Id, m.Text, e.Timestamp)
		}
		return transport.Reply{}, false, nil
	}
	if e, ok := event.(webhook.UnsendEvent); ok {
		g, ok := e.Source.(webhook.GroupSource)
		if !ok || g.GroupId == "" || e.Unsend == nil {
			return transport.Reply{}, false, nil
		}
		return transport.Reply{}, false, b.Store.Unsend(ctx, g.GroupId, e.Unsend.MessageId)
	}
	e, ok := event.(webhook.MessageEvent)
	if !ok || e.Mode != "active" {
		return transport.Reply{}, false, nil
	}
	g, ok := e.Source.(webhook.GroupSource)
	if !ok || g.GroupId == "" {
		return transport.Reply{}, false, nil
	}
	m, ok := e.Message.(webhook.TextMessageContent)
	if !ok {
		return transport.Reply{}, false, nil
	}
	command := strings.TrimSpace(m.Text)
	if skipCollectionText(command) {
		return transport.Reply{}, false, nil
	}
	return transport.Reply{}, false, b.Store.Collect(ctx, g.GroupId, messagecollection.Message{ID: m.Id, EventID: e.WebhookEventId, UserID: g.UserId, Text: m.Text, OccurredAt: time.UnixMilli(e.Timestamp).UTC()})
}

func skipCollectionText(text string) bool {
	if strings.HasPrefix(text, "/") {
		return true
	}
	switch text {
	case "摘要", "上班", "下班", "打卡", "出勤紀錄", "選單", "功能", "關閉摘要", "開啟摘要", "":
		return true
	}
	return false
}
