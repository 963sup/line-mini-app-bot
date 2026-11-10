package line

import (
	"context"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type (
	Reply struct {
		Token   string
		Message messaging_api.MessageInterface
	}
	Handler struct {
		Handle             func(context.Context, webhook.EventInterface) (Reply, bool, error)
		Failure            string
		RetryOnSendFailure bool
	}
)

func CanReply(mode webhook.EventMode, token string, delivery *webhook.DeliveryContext) bool {
	return mode == "active" && token != "" && (delivery == nil || !delivery.IsRedelivery)
}
