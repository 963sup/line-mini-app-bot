package line

import (
	"context"
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestTemplateCommand(t *testing.T) {
	handler := Handler("https://miniapp.line.me/123-example")
	msg := webhook.MessageEvent{Mode: "active", ReplyToken: "token", Message: webhook.TextMessageContent{Text: "物件範本"}}
	reply, handled, err := handler.Handle(context.Background(), msg)
	if err != nil || !handled || reply.Token != "token" {
		t.Fatal(reply, handled, err)
	}
	text, ok := reply.Message.(*messaging_api.TextMessage)
	if !ok || text.Text != "開啟物件範本：https://miniapp.line.me/123-example/object-templates" {
		t.Fatal(reply.Message)
	}
	msg.DeliveryContext = &webhook.DeliveryContext{IsRedelivery: true}
	if _, handled, _ := handler.Handle(context.Background(), msg); handled {
		t.Fatal("replied twice")
	}
}
