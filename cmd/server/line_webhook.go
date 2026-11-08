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
func registerLINEWebhook(mux *http.ServeMux, appURL string, attendance *attendanceBot, summaries *summaryBot) error {
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
	mux.HandleFunc("POST "+path, lineWebhookHandler(secret, bot, appURL, attendance, summaries))
	slog.Info("LINE connectivity prototype enabled", "path", path)
	return nil
}

func lineWebhookHandler(secret string, bot *messaging_api.MessagingApiAPI, appURL string, attendance *attendanceBot, summaries *summaryBot) http.HandlerFunc {
	weather := newWeatherClient()
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
			summaryReply, summaryHandled, summaryErr := summaries.observe(ctx, event)
			if summaryErr != nil {
				http.Error(w, "summary temporarily unavailable", http.StatusServiceUnavailable)
				return
			}
			if summaryHandled {
				if summaryReply.message != nil {
					// Reply is best effort; the durable job and enqueue have already succeeded.
					_, _ = client.ReplyMessage(&messaging_api.ReplyMessageRequest{ReplyToken: summaryReply.token, Messages: []messaging_api.MessageInterface{summaryReply.message}})
				}
				continue
			}
			reply, handled, attendanceErr := attendance.reply(ctx, event)
			if attendanceErr != nil {
				// A retry reuses the persisted event result if commit already succeeded.
				slog.Warn("attendance operation unavailable")
				http.Error(w, "attendance unavailable; retry safely", http.StatusServiceUnavailable)
				return
			}
			if !handled {
				var weatherHandled bool
				reply, weatherHandled = weatherReply(ctx, event, weather)
				if !weatherHandled {
					reply = replyForEvent(event, appURL)
				}
			}
			if reply.message == nil {
				continue
			}
			_, err := client.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: reply.token,
				Messages:   []messaging_api.MessageInterface{reply.message},
			})
			if err != nil {
				if handled {
					http.Error(w, "attendance reply unavailable", http.StatusServiceUnavailable)
					return
				}
				// Raw SDK errors can include request details; log no token or payload.
				slog.Warn("LINE prototype reply failed; not retried")
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
