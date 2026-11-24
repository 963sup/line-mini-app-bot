package membershipline

import (
	"testing"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestInvitationCommandScopeOnRedelivery(t *testing.T) {
	h := Handler(nil)
	for _, redelivery := range []bool{false, true} {
		e := webhook.MessageEvent{
			Mode: "active", WebhookEventId: "event", ReplyToken: "reply",
			Source:          webhook.UserSource{UserId: "owner"},
			Message:         webhook.TextMessageContent{Text: "邀請加入 org"},
			DeliveryContext: &webhook.DeliveryContext{IsRedelivery: redelivery},
		}
		reply, handled, err := h.Handle(t.Context(), e)
		msg, ok := reply.Message.(*messaging_api.TextMessage)
		if err != nil || !handled || !ok || msg.Text != "邀請服務目前無法使用。" || reply.Token != e.ReplyToken {
			t.Fatalf("private, redelivery=%v: %+v, %v, %v", redelivery, reply, handled, err)
		}
		e.Source = webhook.GroupSource{GroupId: "group", UserId: "owner"}
		reply, handled, err = h.Handle(t.Context(), e)
		if err != nil || handled || reply.Message != nil {
			t.Fatalf("group, redelivery=%v: %+v, %v, %v", redelivery, reply, handled, err)
		}
	}
}
