package line

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
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

type SendReply func(context.Context, string, messaging_api.MessageInterface) error

func RegisterWebhook(mux *http.ServeMux, channel, secret string, send SendReply, dispatch Dispatch) {
	path := "/webhooks/line/" + channel
	mux.HandleFunc("POST "+path, WebhookHandler(secret, send, dispatch))
	slog.Info("LINE webhook enabled", "path", path)
}

func WebhookHandler(secret string, send SendReply, dispatch Dispatch) http.HandlerFunc {
	if err := dispatch.validate(); err != nil {
		panic(err)
	}
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
		for _, event := range callback.Events {
			if e, ok := event.(webhook.MessageEvent); ok && e.Mode == "active" {
				if g, ok := e.Source.(webhook.GroupSource); ok && g.GroupId != "" {
					if sticker, ok := e.Message.(webhook.StickerMessageContent); ok {
						slog.Info("LINE group sticker received", "packageId", sticker.PackageId, "stickerId", sticker.StickerId, "resourceType", string(sticker.StickerResourceType))
					}
				}
			}
			if handled, ok := runCommands(ctx, w, send, event, dispatch.EarlyCommands); !ok {
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
			if _, ok := runCommands(ctx, w, send, event, dispatch.Commands); !ok {
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (d Dispatch) validate() error {
	commands, actions := map[string]string{}, map[string]string{}
	for _, phase := range [][]Handler{d.EarlyCommands, d.Commands} {
		for _, h := range phase {
			for _, name := range h.Commands {
				key := strings.ToLower(strings.TrimSpace(name))
				if prev, ok := commands[key]; ok {
					return fmt.Errorf("LINE command %q registered by %s and %s", name, prev, h.Owner)
				}
				commands[key] = h.Owner
			}
			for _, key := range h.Postbacks {
				if prev, ok := actions[key]; ok {
					return fmt.Errorf("LINE postback %q registered by %s and %s", key, prev, h.Owner)
				}
				actions[key] = h.Owner
			}
		}
	}
	return nil
}

// runCommands reports whether a command claimed the event and whether the
// webhook may continue processing the batch.
func runCommands(ctx context.Context, w http.ResponseWriter, send SendReply, event webhook.EventInterface, commands []Handler) (handled, ok bool) {
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
			sendErr := send(ctx, reply.Token, reply.Message)
			if sendErr != nil && command.RetryOnSendFailure && !spentRedeliveryToken(event, sendErr) {
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
func spentRedeliveryToken(event webhook.EventInterface, err error) bool {
	var delivery *webhook.DeliveryContext
	switch e := event.(type) {
	case webhook.MessageEvent:
		delivery = e.DeliveryContext
	case webhook.PostbackEvent:
		delivery = e.DeliveryContext
	}
	return delivery != nil && delivery.IsRedelivery && errors.Is(err, lineplatform.ErrInvalidReplyToken)
}
