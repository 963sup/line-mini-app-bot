package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/lineui"
	"github.com/963sup/line-mini-app-bot/internal/weather"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type weatherLookup interface {
	Lookup(context.Context, float64, float64, string) (weather.Forecast, error)
}

func newWeatherClient() weatherLookup {
	key := os.Getenv("CWA_API_KEY")
	if key == "" {
		return nil
	}
	return &weather.Client{APIKey: key, HTTP: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func weatherReply(ctx context.Context, event webhook.EventInterface, client weatherLookup) (lineReply, bool) {
	e, ok := event.(webhook.MessageEvent)
	if !ok || !canReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
		return lineReply{}, false
	}
	g, ok := e.Source.(webhook.GroupSource)
	if !ok || g.GroupId == "" {
		return lineReply{}, false
	}
	m, ok := e.Message.(webhook.LocationMessageContent)
	if !ok {
		return lineReply{}, false
	}
	reply := lineReply{token: e.ReplyToken}
	if client == nil {
		reply.message = &messaging_api.TextMessage{Text: "天氣查詢尚未設定。"}
		return reply, true
	}
	f, err := client.Lookup(ctx, m.Latitude, m.Longitude, m.Address)
	if err != nil {
		text := "暫時無法取得天氣，請稍後再分享位置。"
		if errors.Is(err, weather.ErrOutsideCoverage) {
			text = "此位置不在目前支援的臺灣區域預報範圍內。"
		}
		reply.message = &messaging_api.TextMessage{Text: text}
	} else {
		reply.message = lineui.Weather(f)
	}
	return reply, true
}
