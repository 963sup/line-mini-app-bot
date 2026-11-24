package line

import (
	"context"
	"testing"
	"time"

	dto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type creationStub struct{ calls int }

func (s *creationStub) Conversation(_ context.Context, user, event, text string, _ time.Time) (*dto.CreationReply, error) {
	s.calls++
	return &dto.CreationReply{Text: "已建立", Kind: "organization", ID: "org1"}, nil
}

func TestCreationRepliesOnlyToPrivateMessages(t *testing.T) {
	store := &creationStub{}
	handler := CreationHandler(store, "https://miniapp.line.me/example")
	m := webhook.MessageEvent{Mode: "active", WebhookEventId: "event", Timestamp: time.Now().UnixMilli(), ReplyToken: "token", Source: webhook.GroupSource{GroupId: "g"}, Message: webhook.TextMessageContent{Text: "建立組織"}}
	if _, ok, e := handler.Handle(t.Context(), m); ok || e != nil || store.calls != 0 {
		t.Fatal("group reached creation store")
	}
	m.Source = webhook.UserSource{UserId: "user"}
	m.DeliveryContext = &webhook.DeliveryContext{IsRedelivery: true}
	reply, ok, e := handler.Handle(t.Context(), m)
	if !ok || e != nil || store.calls != 1 {
		t.Fatal("lost redelivery", e)
	}
	message := reply.Message.(*messaging_api.TemplateMessage)
	action := message.Template.(*messaging_api.ButtonsTemplate).Actions[0].(*messaging_api.UriAction)
	if action.Uri != "https://miniapp.line.me/example?organization=org1" || action.Label != "開啟組織" {
		t.Fatal(action)
	}
}
