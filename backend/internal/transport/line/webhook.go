package line

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// Observer records a fact without claiming the event or sending a reply.
type Observer struct {
	Observe func(context.Context, webhook.EventInterface, *GroupEvent) error
	Failure string
}

// Dispatch preserves early command precedence while explicitly separating
// observation from command handling.
type Dispatch struct {
	EarlyCommands []Handler
	Observers     []Observer
	Commands      []Handler
}

func WebhookHandler(secret string, bot *messaging_api.MessagingApiAPI, dispatch Dispatch) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		callback, err := webhook.ParseRequest(secret, r)
		if err != nil {
			status := http.StatusBadRequest
			var sizeError *http.MaxBytesError
			switch {
			case errors.As(err, &sizeError):
				status = http.StatusRequestEntityTooLarge
			case errors.Is(err, webhook.ErrInvalidSignature):
				status = http.StatusUnauthorized
			}
			http.Error(w, "invalid webhook request", status)
			return
		}
		// One deadline bounds the whole batch, not five seconds per event.
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		// SDK WithContext mutates its receiver; never mutate the shared client.
		client := *bot
		client.WithContext(ctx)
		for _, event := range callback.Events {
			if e, ok := event.(webhook.MessageEvent); ok && e.Mode == "active" {
				if g, ok := e.Source.(webhook.GroupSource); ok && g.GroupId != "" {
					if sticker, ok := e.Message.(webhook.StickerMessageContent); ok {
						slog.Info("LINE group sticker received", "packageId", sticker.PackageId, "stickerId", sticker.StickerId, "resourceType", string(sticker.StickerResourceType))
					}
				}
			}
			if handled, ok := runCommands(ctx, w, &client, event, dispatch.EarlyCommands); !ok {
				return
			} else if handled {
				continue
			}
			var group *GroupEvent
			if len(dispatch.Observers) > 0 {
				if decoded, ok := DecodeGroupEvent(event); ok {
					group = &decoded
				}
			}
			for _, observer := range dispatch.Observers {
				if err := observer.Observe(ctx, event, group); err != nil {
					http.Error(w, observer.Failure, http.StatusServiceUnavailable)
					return
				}
			}
			if _, ok := runCommands(ctx, w, &client, event, dispatch.Commands); !ok {
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

// runCommands reports whether a command claimed the event and whether the
// webhook may continue processing the batch.
func runCommands(ctx context.Context, w http.ResponseWriter, client *messaging_api.MessagingApiAPI, event webhook.EventInterface, commands []Handler) (handled, ok bool) {
	for _, command := range commands {
		reply, handled, err := command.Handle(ctx, event)
		if err != nil {
			http.Error(w, command.Failure, http.StatusServiceUnavailable)
			return false, false
		}
		if !handled {
			continue
		}
		if reply.Message != nil {
			response, _, sendErr := client.ReplyMessageWithHttpInfo(&messaging_api.ReplyMessageRequest{ReplyToken: reply.Token, Messages: []messaging_api.MessageInterface{reply.Message}})
			if sendErr != nil && command.RetryOnSendFailure && !spentRedeliveryToken(event, response) {
				http.Error(w, "LINE reply unavailable", http.StatusServiceUnavailable)
				return true, false
			}
			if sendErr != nil {
				slog.Warn("LINE reply failed; not retried")
			}
		}
		return true, true
	}
	return false, true
}

// A token rejected on redelivery cannot recover by retrying the same reply.
func spentRedeliveryToken(event webhook.EventInterface, response *http.Response) bool {
	var delivery *webhook.DeliveryContext
	switch e := event.(type) {
	case webhook.MessageEvent:
		delivery = e.DeliveryContext
	case webhook.PostbackEvent:
		delivery = e.DeliveryContext
	}
	if delivery == nil || !delivery.IsRedelivery || response == nil || response.StatusCode != http.StatusBadRequest {
		return false
	}
	var rejection struct {
		Message string `json:"message"`
	}
	return json.NewDecoder(response.Body).Decode(&rejection) == nil && rejection.Message == "Invalid reply token"
}
