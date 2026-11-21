// Package membershipline adapts verified private LINE messages to Membership.
package membershipline

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func Handler(service *application.Service) transport.Handler {
	return transport.Handler{
		Failure: "membership temporarily unavailable", RetryOnSendFailure: true,
		Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
			m, ok := event.(webhook.MessageEvent)
			if !ok || m.Mode != "active" || m.ReplyToken == "" || m.WebhookEventId == "" {
				return transport.Reply{}, false, nil
			}
			source, private := m.Source.(webhook.UserSource)
			message, text := m.Message.(webhook.TextMessageContent)
			if !private || !text || source.UserId == "" {
				return transport.Reply{}, false, nil
			}
			args := strings.Fields(message.Text)
			if len(args) == 0 {
				return transport.Reply{}, false, nil
			}
			switch args[0] {
			case "邀請加入", "接受邀請", "撤銷邀請":
			default:
				return transport.Reply{}, false, nil
			}
			reply := func(text string) (transport.Reply, bool, error) {
				return transport.Reply{Token: m.ReplyToken, Message: &messaging_api.TextMessage{Text: text}}, true, nil
			}
			if len(args) != 2 {
				return reply("請輸入「邀請加入 組織ID」、「接受邀請 邀請碼」或「撤銷邀請 邀請碼」。")
			}
			if service == nil {
				return reply("邀請服務目前無法使用。")
			}
			if args[0] != "邀請加入" {
				raw, err := base64.RawURLEncoding.DecodeString(args[1])
				if err != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != args[1] {
					return reply("邀請碼無效。")
				}
			}
			var err error
			switch args[0] {
			case "邀請加入":
				var code string
				code, err = service.Invite(ctx, source.UserId, m.WebhookEventId, args[1])
				if err == nil {
					return reply("將下面這行私訊給對方，請對方傳給官方帳號。邀請 24 小時有效，任何拿到碼的人都可能先使用。\n\n接受邀請 " + code)
				}
			case "接受邀請":
				err = service.Accept(ctx, source.UserId, m.WebhookEventId, args[1])
				if err == nil {
					return reply("已加入組織。")
				}
			case "撤銷邀請":
				err = service.Revoke(ctx, source.UserId, args[1])
				if err == nil {
					return reply("已撤銷邀請。")
				}
			}
			switch {
			case errors.Is(err, application.ErrAlreadyMember):
				return reply("你已經是這個組織的成員。")
			case errors.Is(err, application.ErrForbidden), errors.Is(err, application.ErrUnavailableInvitation):
				return reply("這筆邀請無法操作，請向組織擁有者確認。")
			default:
				return transport.Reply{}, true, err
			}
		},
	}
}
