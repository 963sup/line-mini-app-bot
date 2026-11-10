package line

import (
	"context"
	"errors"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/summary"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type SummaryDelivery struct {
	Bot *messaging_api.MessagingApiAPI
}

func (d SummaryDelivery) Send(ctx context.Context, group, text, retry string) error {
	client := *d.Bot
	client.WithContext(ctx)
	res, _, e := client.PushMessageWithHttpInfo(&messaging_api.PushMessageRequest{To: group, Messages: []messaging_api.MessageInterface{&messaging_api.TextMessage{Text: text}}}, retry)
	// LINE's 409 means this retry key was already accepted, not a new delivery.
	if res != nil && res.StatusCode == http.StatusConflict {
		return nil
	}
	if e != nil {
		if errors.Is(e, context.Canceled) {
			return context.Canceled
		}
		if errors.Is(e, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return summary.ErrDeliveryUnavailable
	}
	return nil
}
