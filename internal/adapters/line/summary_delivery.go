package line

import (
	"context"
	"net/http"

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
	return e
}
