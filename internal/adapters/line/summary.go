package line

import (
	"context"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/summary"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Summary struct {
	Store summary.Store
	Queue summary.Queue
}

// observe is independent of command replies: ordinary text remains silent.
func (b *Summary) observe(ctx context.Context, event webhook.EventInterface) (lineReply, bool, error) {
	if b == nil {
		return lineReply{}, false, nil
	}
	if e, ok := event.(webhook.MessageEditedEvent); ok {
		g, group := e.Source.(webhook.GroupSource)
		m, text := e.Message.(webhook.TextMessageContent)
		if editor, supported := b.Store.(interface {
			Edit(context.Context, string, string, string, int64) error
		}); supported && group && text && g.GroupId != "" && e.Mode == "active" {
			return lineReply{}, false, editor.Edit(ctx, g.GroupId, m.Id, m.Text, e.Timestamp)
		}
		return lineReply{}, false, nil
	}
	if e, ok := event.(webhook.UnsendEvent); ok {
		g, ok := e.Source.(webhook.GroupSource)
		if !ok || g.GroupId == "" || e.Unsend == nil {
			return lineReply{}, false, nil
		}
		return lineReply{}, false, b.Store.Unsend(ctx, g.GroupId, e.Unsend.MessageId)
	}
	e, ok := event.(webhook.MessageEvent)
	if !ok || e.Mode != "active" {
		return lineReply{}, false, nil
	}
	g, ok := e.Source.(webhook.GroupSource)
	if !ok || g.GroupId == "" {
		return lineReply{}, false, nil
	}
	m, ok := e.Message.(webhook.TextMessageContent)
	if !ok {
		return lineReply{}, false, nil
	}
	command := strings.TrimSpace(m.Text)
	if command == "摘要" {
		if time.Since(time.UnixMilli(e.Timestamp)) > 24*time.Hour || time.UnixMilli(e.Timestamp).After(time.Now().Add(time.Minute)) {
			return lineReply{}, true, nil
		}
		if g.UserId == "" || e.WebhookEventId == "" || e.ReplyToken == "" {
			return lineReply{}, true, nil
		}
		j, err := b.Store.Command(ctx, g.GroupId, e.WebhookEventId, command, e.Timestamp)
		if err != nil {
			return lineReply{}, true, err
		}
		if j.Status == "done" {
			return lineReply{}, true, nil
		}
		if err := b.Queue.Enqueue(ctx, j); err != nil {
			return lineReply{}, true, err
		}
		text := "正在整理最近24小時的群聊（最多100則）。"
		return lineReply{token: e.ReplyToken, message: &messaging_api.TextMessage{Text: text}}, true, nil
	}
	if skipSummaryText(command) {
		return lineReply{}, false, nil
	}
	return lineReply{}, false, b.Store.Collect(ctx, g.GroupId, summary.Message{ID: m.Id, EventID: e.WebhookEventId, UserID: g.UserId, Text: m.Text, OccurredAt: time.UnixMilli(e.Timestamp).UTC()})
}

func skipSummaryText(text string) bool {
	if strings.HasPrefix(text, "/") {
		return true
	}
	switch text {
	case "上班", "下班", "打卡", "出勤紀錄", "選單", "功能", "關閉摘要", "開啟摘要", "":
		return true
	}
	return false
}
