package line

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

var (
	ErrInvalidReplyToken = errors.New("invalid LINE reply token")
	errSend              = errors.New("LINE message delivery unavailable")
)

// Messaging shares the SDK's HTTP connection pool across webhook and worker requests.
type Messaging struct {
	api *messaging_api.MessagingApiAPI
}

func NewMessaging(token string, httpClient *http.Client) (*Messaging, error) {
	api, err := messaging_api.NewMessagingApiAPI(token, messaging_api.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &Messaging{api: api}, nil
}

func (m *Messaging) request(ctx context.Context) *messaging_api.MessagingApiAPI {
	// WithContext mutates its receiver.
	client := *m.api
	return client.WithContext(ctx)
}

func (m *Messaging) Reply(ctx context.Context, token string, message messaging_api.MessageInterface) error {
	response, _, err := m.request(ctx).ReplyMessageWithHttpInfo(&messaging_api.ReplyMessageRequest{ReplyToken: token, Messages: []messaging_api.MessageInterface{message}})
	if err != nil && response != nil && response.StatusCode == http.StatusBadRequest && response.Body != nil {
		var rejection struct{ Message string }
		if json.NewDecoder(response.Body).Decode(&rejection) == nil && rejection.Message == "Invalid reply token" {
			return ErrInvalidReplyToken
		}
	}
	return sendError(err)
}

func (m *Messaging) PushText(ctx context.Context, group, text, retryKey string) error {
	response, _, err := m.request(ctx).PushMessageWithHttpInfo(&messaging_api.PushMessageRequest{To: group, Messages: []messaging_api.MessageInterface{&messaging_api.TextMessage{Text: text}}}, retryKey)
	if response != nil && response.StatusCode == http.StatusConflict {
		return nil
	}
	return sendError(err)
}

func sendError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	default:
		return errSend
	}
}
