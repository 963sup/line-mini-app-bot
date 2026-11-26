package line

import (
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func TestPublishedAttendanceActionsDispatch(t *testing.T) {
	h := navigation.Menu{PrivateActions: NavigationActions()}.Handler("", "", nil)
	for _, menu := range navigation.AttendanceRichMenus() {
		data := menu.Areas[2].Action.(*messaging_api.PostbackAction).Data
		e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: data}}
		reply, handled, err := h.Handle(t.Context(), e)
		message, ok := reply.Message.(*messaging_api.TextMessage)
		if err != nil || !handled || !ok || message.Text == "" {
			t.Fatalf("private postback %s: %+v, %v", data, reply, err)
		}
		e.Source = webhook.GroupSource{GroupId: "group"}
		group, _, err := h.Handle(t.Context(), e)
		if err != nil || group.Message != nil {
			t.Fatalf("group postback %s: %+v, %v", data, group, err)
		}
		command := "/" + strings.ToUpper(strings.TrimPrefix(data, "action="))
		text := webhook.MessageEvent{Mode: "active", ReplyToken: "reply", Source: e.Source, Message: webhook.TextMessageContent{Text: command}}
		group, handled, err = h.Handle(t.Context(), text)
		if err != nil || handled || group.Message != nil {
			t.Fatalf("group command %s: %+v, %v", command, group, err)
		}
	}
}

func TestLegacyTransferActionDispatchesAsSupport(t *testing.T) {
	h := navigation.Menu{PrivateActions: NavigationActions()}.Handler("", "", nil)
	e := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: "action=transfer"}}
	reply, handled, err := h.Handle(t.Context(), e)
	message, ok := reply.Message.(*messaging_api.TextMessage)
	if err != nil || !handled || !ok || !strings.Contains(message.Text, "支援") {
		t.Fatalf("legacy transfer postback: %+v, %v", reply, err)
	}
}
