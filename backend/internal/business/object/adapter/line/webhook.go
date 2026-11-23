package line

import (
	"context"
	"net/url"
	"strings"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// Handler opens the existing template library; authorization stays in the Mini App.
func Handler(appURL string) transport.Handler {
	return transport.Handler{Handle: func(_ context.Context, e webhook.EventInterface) (transport.Reply, bool, error) {
		msg, ok := e.(webhook.MessageEvent)
		if !ok || !transport.CanReply(msg.Mode, msg.ReplyToken, msg.DeliveryContext) {
			return transport.Reply{}, false, nil
		}
		t, ok := msg.Message.(webhook.TextMessageContent)
		if !ok || strings.TrimSpace(t.Text) != "物件範本" {
			return transport.Reply{}, false, nil
		}
		target, err := url.JoinPath(appURL, "object-templates")
		if err != nil || appURL == "" {
			return transport.Reply{}, false, nil
		}
		return transport.Reply{Token: msg.ReplyToken, Message: &messaging_api.TextMessage{Text: "開啟物件範本：" + target}}, true, nil
	}}
}
