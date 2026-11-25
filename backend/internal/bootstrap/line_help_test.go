package bootstrap

import (
	"slices"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestHelpMenuViaWebhookHandler(t *testing.T) {
	const mobile, desktop = "https://miniapp.line.me/example", "https://example.org/desktop"
	handler := navigation.Menu{QuickMenu: quickMenu}.Handler(mobile, desktop, nil)
	for _, tc := range []struct {
		name, input string
		source      webhook.SourceInterface
		labels      []string
	}{
		{"private", "/help", webhook.UserSource{UserId: "u"}, []string{"我的出勤", "建立組織", "建立企業", "空間", "MINI App"}},
		{"private fallback", "unknown", webhook.UserSource{UserId: "u"}, []string{"我的出勤", "建立組織", "建立企業", "空間", "MINI App"}},
		{"group", "/help", webhook.GroupSource{GroupId: "g"}, []string{"打卡", "日誌", "出勤", "請假", "補卡", "更正", "摘要", "MINI App"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event := webhook.MessageEvent{Source: tc.source, Mode: "active", ReplyToken: "reply", Message: webhook.TextMessageContent{Text: tc.input}}
			reply, handled, err := handler.Handle(t.Context(), event)
			if err != nil || !handled || reply.Token != "reply" {
				t.Fatalf("handle: %+v %v", reply, err)
			}
			message := reply.Message.(*messaging_api.TextMessage)
			var labels []string
			for _, item := range message.QuickReply.Items {
				switch a := item.Action.(type) {
				case *messaging_api.MessageAction:
					labels = append(labels, a.Label)
				case *messaging_api.PostbackAction:
					labels = append(labels, a.Label)
				case *messaging_api.UriAction:
					labels = append(labels, a.Label)
				}
			}
			if !slices.Equal(labels, tc.labels) {
				t.Fatalf("buttons %v, want %v", labels, tc.labels)
			}
		})
	}
}
