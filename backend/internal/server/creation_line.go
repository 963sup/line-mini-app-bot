package server

import (
	"context"
	"net/url"
	"strings"
	"time"

	orgdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type creationConversation interface {
	Conversation(context.Context, string, string, string, time.Time) (*orgdb.CreationReply, error)
}

func creationHandler(store creationConversation, appURL string) transport.Handler {
	return transport.Handler{Failure: "creation temporarily unavailable", RetryOnSendFailure: true, Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
		m, ok := event.(webhook.MessageEvent)
		if !ok || m.Mode != "active" || m.ReplyToken == "" || m.WebhookEventId == "" {
			return transport.Reply{}, false, nil
		}
		user, private := m.Source.(webhook.UserSource)
		text, ok := m.Message.(webhook.TextMessageContent)
		if !private || user.UserId == "" || !ok {
			return transport.Reply{}, false, nil
		}
		if store == nil {
			command := strings.TrimSpace(text.Text)
			if command != "建立企業" && command != "建立組織" {
				return transport.Reply{}, false, nil
			}
			return transport.Reply{Token: m.ReplyToken, Message: &messaging_api.TextMessage{Text: "建立服務目前無法使用，請稍後再試。"}}, true, nil
		}
		res, e := store.Conversation(ctx, user.UserId, m.WebhookEventId, text.Text, time.UnixMilli(m.Timestamp))
		if e != nil || res == nil {
			return transport.Reply{}, false, e
		}
		message := &messaging_api.TextMessage{Text: res.Text}
		if res.ID != "" {
			link, e := url.Parse(appURL)
			if e == nil && link.Scheme == "https" && link.Host != "" {
				q := link.Query()
				q.Set(res.Kind, res.ID)
				link.RawQuery = q.Encode()
				label := "開啟組織"
				if res.Kind == "enterprise" {
					label = "開啟企業"
				}
				message.QuickReply = &messaging_api.QuickReply{Items: []messaging_api.QuickReplyItem{{Type: "action", Action: &messaging_api.UriAction{Label: label, Uri: link.String()}}}}
			}
		}
		return transport.Reply{Token: m.ReplyToken, Message: message}, true, nil
	}}
}
