package line

import (
	"context"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/summary/summarization"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Commands struct {
	Store summarization.CommandStore
	Queue summarization.Queue
}

func (b *Commands) Reply(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
	if b == nil {
		return transport.Reply{}, false, nil
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
	if command == "摘要" {
		if time.Since(time.UnixMilli(e.Timestamp)) > 24*time.Hour || time.UnixMilli(e.Timestamp).After(time.Now().Add(time.Minute)) {
			return transport.Reply{}, true, nil
		}
		if g.UserId == "" || e.WebhookEventId == "" || e.ReplyToken == "" {
			return transport.Reply{}, true, nil
		}
		j, err := b.Store.Command(ctx, g.GroupId, e.WebhookEventId, command, e.Timestamp)
		if err != nil {
			return transport.Reply{}, true, err
		}
		if j.Status == "done" {
			return transport.Reply{}, true, nil
		}
		if err := b.Queue.Enqueue(ctx, j); err != nil {
			return transport.Reply{}, true, err
		}
		text := "正在整理最近24小時的群聊（最多100則）。"
		return transport.Reply{Token: e.ReplyToken, Message: &messaging_api.TextMessage{Text: text}}, true, nil
	}
	return transport.Reply{}, false, nil
}
