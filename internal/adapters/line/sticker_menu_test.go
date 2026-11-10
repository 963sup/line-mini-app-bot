package line

import (
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestStickerMenuExactMatch(t *testing.T) {
	t.Setenv("LINE_MENU_STICKER_PACKAGE_ID", "123")
	t.Setenv("LINE_MENU_STICKER_ID", "456")
	makeEvent := func(pkg, id string) webhook.MessageEvent {
		return webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Message: webhook.StickerMessageContent{PackageId: pkg, StickerId: id}}
	}
	if _, ok := replyForEvent(makeEvent("123", "456"), "").message.(*messaging_api.TextMessage); !ok {
		t.Fatal("matching sticker must show quick menu")
	}
	for _, tc := range [][2]string{{"123", "999"}, {"999", "456"}, {"", ""}} {
		if replyForEvent(makeEvent(tc[0], tc[1]), "").message != nil {
			t.Fatal("unmatched sticker must stay silent")
		}
	}
	t.Setenv("LINE_MENU_STICKER_ID", "")
	if replyForEvent(makeEvent("123", "456"), "").message != nil {
		t.Fatal("unconfigured sticker must stay silent")
	}
}
