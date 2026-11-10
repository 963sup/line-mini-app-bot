package line

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// RegisterWebhook enables the connectivity prototype only when configured.
func RegisterWebhook(mux *http.ServeMux, handlers []Handler) (func(context.Context, string) string, error) {
	channelID := os.Getenv("LINE_MESSAGING_API_CHANNEL_ID")
	secret := os.Getenv("LINE_MESSAGING_API_CHANNEL_SECRET")
	token := os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	configurationAbsent := channelID == "" && secret == "" && token == ""
	if configurationAbsent {
		slog.Info("LINE webhook disabled: Messaging API configuration is absent")
		return nil, nil
	}
	settings := []struct{ name, value string }{
		{"LINE_MESSAGING_API_CHANNEL_ID", channelID},
		{"LINE_MESSAGING_API_CHANNEL_SECRET", secret},
		{"LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN", token},
	}
	for _, setting := range settings {
		if strings.TrimSpace(setting.value) == "" {
			return nil, fmt.Errorf("missing configuration: %s", setting.name)
		}
	}
	for _, digit := range channelID {
		if digit < '0' || digit > '9' {
			return nil, errors.New("LINE_MESSAGING_API_CHANNEL_ID must contain only digits")
		}
	}
	bot, err := messaging_api.NewMessagingApiAPI(token,
		messaging_api.WithHTTPClient(&http.Client{Timeout: 5 * time.Second}))
	if err != nil {
		return nil, errors.New("cannot initialize LINE Messaging API client")
	}
	lookup := func(ctx context.Context, group string) string {
		lookupCtx, cancel := context.WithTimeout(ctx, 400*time.Millisecond)
		defer cancel()
		client := *bot
		client.WithContext(lookupCtx)
		summary, err := client.GetGroupSummary(group)
		if err != nil || summary == nil {
			return ""
		}
		return summary.GroupName
	}
	path := "/webhooks/line/" + channelID
	mux.HandleFunc("POST "+path, WebhookHandler(secret, bot, handlers))
	slog.Info("LINE connectivity prototype enabled", "path", path)
	return lookup, nil
}

func WebhookHandler(secret string, bot *messaging_api.MessagingApiAPI, handlers []Handler) http.HandlerFunc {
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
			for _, handler := range handlers {
				reply, handled, handleErr := handler.Handle(ctx, event)
				if handleErr != nil {
					http.Error(w, handler.Failure, http.StatusServiceUnavailable)
					return
				}
				if !handled {
					continue
				}
				if reply.Message != nil {
					_, sendErr := client.ReplyMessage(&messaging_api.ReplyMessageRequest{ReplyToken: reply.Token, Messages: []messaging_api.MessageInterface{reply.Message}})
					if sendErr != nil && handler.RetryOnSendFailure {
						http.Error(w, "attendance reply unavailable", http.StatusServiceUnavailable)
						return
					}
					if sendErr != nil {
						slog.Warn("LINE reply failed; not retried")
					}
				}
				break
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
