package navigation

import (
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestStickerMenuExactMatch(t *testing.T) {
	trigger := Menu{StickerPackageID: "123", StickerID: "456"}
	makeEvent := func(pkg, id string) webhook.MessageEvent {
		return webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: webhook.GroupSource{GroupId: "group"}, Message: webhook.StickerMessageContent{PackageId: pkg, StickerId: id}}
	}
	if _, ok := trigger.ReplyForEvent(makeEvent("123", "456"), "", "").Message.(*messaging_api.TextMessage); !ok {
		t.Fatal("matching sticker must show quick menu")
	}
	for _, tc := range [][2]string{{"123", "999"}, {"999", "456"}, {"", ""}} {
		if trigger.ReplyForEvent(makeEvent(tc[0], tc[1]), "", "").Message != nil {
			t.Fatal("unmatched sticker must stay silent")
		}
	}
	trigger.StickerID = ""
	if trigger.ReplyForEvent(makeEvent("123", "456"), "", "").Message != nil {
		t.Fatal("unconfigured sticker must stay silent")
	}
}
