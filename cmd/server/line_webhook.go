package main

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

// registerLINEWebhook enables the connectivity prototype only when configured.
func registerLINEWebhook(mux *http.ServeMux) error {
	channelID := os.Getenv("LINE_MESSAGING_API_CHANNEL_ID")
	secret := os.Getenv("LINE_MESSAGING_API_CHANNEL_SECRET")
	token := os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	configurationAbsent := channelID == "" && secret == "" && token == ""
	if configurationAbsent {
		slog.Info("LINE webhook disabled: Messaging API configuration is absent")
		return nil
	}
	settings := []struct{ name, value string }{
		{"LINE_MESSAGING_API_CHANNEL_ID", channelID},
		{"LINE_MESSAGING_API_CHANNEL_SECRET", secret},
		{"LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN", token},
	}
	for _, setting := range settings {
		if strings.TrimSpace(setting.value) == "" {
			return fmt.Errorf("missing configuration: %s", setting.name)
		}
	}
	for _, digit := range channelID {
		if digit < '0' || digit > '9' {
			return errors.New("LINE_MESSAGING_API_CHANNEL_ID must contain only digits")
		}
	}
	bot, err := messaging_api.NewMessagingApiAPI(token,
		messaging_api.WithHTTPClient(&http.Client{Timeout: 5 * time.Second}))
	if err != nil {
		return errors.New("cannot initialize LINE Messaging API client")
	}
	path := "/webhooks/line/" + channelID
	mux.HandleFunc("POST "+path, lineWebhookHandler(secret, bot))
	slog.Info("LINE connectivity prototype enabled", "path", path)
	return nil
}

func lineWebhookHandler(secret string, bot *messaging_api.MessagingApiAPI) http.HandlerFunc {
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
			message, ok := event.(webhook.MessageEvent)
			if !ok {
				continue
			}
			if message.Mode != "active" || message.ReplyToken == "" {
				continue
			}
			if _, ok := message.Source.(webhook.UserSource); !ok {
				continue
			}
			if _, ok := message.Message.(webhook.TextMessageContent); !ok {
				continue
			}
			// Do not deliberately retry a short-lived reply in this prototype.
			if message.DeliveryContext != nil && message.DeliveryContext.IsRedelivery {
				continue
			}
			_, err := client.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: message.ReplyToken,
				Messages: []messaging_api.MessageInterface{
					messaging_api.TextMessage{Text: "LINE 連線已接通，目前尚未提供出勤功能。"},
				},
			})
			if err != nil {
				// Raw SDK errors can include request details; log no token or payload.
				slog.Warn("LINE prototype reply failed; not retried")
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
