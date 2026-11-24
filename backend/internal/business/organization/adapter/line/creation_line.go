package line

import (
	"context"
	"net/url"
	"strings"
	"time"

	dto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

const (
	CreateOrganizationCommand = "建立組織"
	CreateEnterpriseCommand   = "建立企業"
)

type Conversation interface {
	Conversation(context.Context, string, string, string, time.Time) (*dto.CreationReply, error)
}

func CreationHandler(store Conversation, appURL string) transport.Handler {
	return transport.Handler{Owner: "organization", Commands: []string{CreateEnterpriseCommand, CreateOrganizationCommand}, Failure: "creation temporarily unavailable", RetryOnSendFailure: true, Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
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
			if command != CreateEnterpriseCommand && command != CreateOrganizationCommand {
				return transport.Reply{}, false, nil
			}
			return transport.Reply{Token: m.ReplyToken, Message: &messaging_api.TextMessage{Text: "建立服務目前無法使用，請稍後再試。"}}, true, nil
		}
		res, e := store.Conversation(ctx, user.UserId, m.WebhookEventId, text.Text, time.UnixMilli(m.Timestamp))
		if e != nil || res == nil {
			return transport.Reply{}, false, e
		}
		var message messaging_api.MessageInterface = &messaging_api.TextMessage{Text: res.Text}
		if res.ID != "" {
			resultText := res.Text
			if res.Kind == "organization" {
				resultText += "\n\n邀請成員時，請傳送：\n邀請加入 " + res.ID
			}
			link, e := url.Parse(appURL)
			if e == nil && link.Scheme == "https" && link.Host != "" {
				q := link.Query()
				q.Set(res.Kind, res.ID)
				link.RawQuery = q.Encode()
				label := "開啟組織"
				if res.Kind == "enterprise" {
					label = "開啟企業"
				}
				message = &messaging_api.TemplateMessage{AltText: label, Template: &messaging_api.ButtonsTemplate{Text: resultText, Actions: []messaging_api.ActionInterface{&messaging_api.UriAction{Label: label, Uri: link.String()}}}}
			} else {
				message = &messaging_api.TextMessage{Text: resultText}
			}
		}
		return transport.Reply{Token: m.ReplyToken, Message: message}, true, nil
	}}
}
