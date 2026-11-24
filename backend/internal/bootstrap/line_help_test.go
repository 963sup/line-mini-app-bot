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
	handler := navigation.Menu{Help: helpMenu}.Handler(mobile, desktop, nil)
	for _, tc := range []struct {
		name, input string
		source      webhook.SourceInterface
		labels      []string
	}{
		{"private", "/help", webhook.UserSource{UserId: "u"}, []string{"出勤紀錄", "空間", "開啟 MINI App"}},
		{"private fallback", "unknown", webhook.UserSource{UserId: "u"}, []string{"出勤紀錄", "空間", "開啟 MINI App"}},
		{"group", "/help", webhook.GroupSource{GroupId: "g"}, []string{"打卡", "日誌", "出勤", "請假登記", "補卡申請", "開啟 MINI App"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event := webhook.MessageEvent{Source: tc.source, Mode: "active", ReplyToken: "reply", Message: webhook.TextMessageContent{Text: tc.input}}
			reply, handled, err := handler.Handle(t.Context(), event)
			if err != nil || !handled || reply.Token != "reply" {
				t.Fatalf("handle: %+v %v", reply, err)
			}
			card := reply.Message.(*messaging_api.FlexMessage)
			var labels []string
			for _, component := range card.Contents.(*messaging_api.FlexBubble).Footer.Contents {
				switch a := component.(*messaging_api.FlexButton).Action.(type) {
				case *messaging_api.MessageAction:
					labels = append(labels, a.Label)
				case *messaging_api.PostbackAction:
					if a.Data != "action=attendance-history" {
						t.Fatal(a.Data)
					}
					labels = append(labels, a.Label)
				case *messaging_api.UriAction:
					if a.Label == "開啟 MINI App" && (a.Uri != mobile || a.AltUri == nil || a.AltUri.Desktop != desktop) {
						t.Fatal(a)
					}
					labels = append(labels, a.Label)
				}
			}
			if !slices.Equal(labels, tc.labels) {
				t.Fatalf("buttons %v, want %v", labels, tc.labels)
			}
		})
	}
}
