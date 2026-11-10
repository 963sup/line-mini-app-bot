package line

import (
	"context"
	"errors"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"

	"github.com/963sup/line-mini-app-bot/internal/weather"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type ForecastLookup interface {
	Lookup(context.Context, float64, float64, string) (weather.Forecast, error)
}

func ReplyForEvent(ctx context.Context, event webhook.EventInterface, client ForecastLookup) (transport.Reply, bool) {
	e, ok := event.(webhook.MessageEvent)
	if !ok || !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
		return transport.Reply{}, false
	}
	g, ok := e.Source.(webhook.GroupSource)
	if !ok || g.GroupId == "" {
		return transport.Reply{}, false
	}
	m, ok := e.Message.(webhook.LocationMessageContent)
	if !ok {
		return transport.Reply{}, false
	}
	reply := transport.Reply{Token: e.ReplyToken}
	if client == nil {
		reply.Message = &messaging_api.TextMessage{Text: "天氣查詢尚未設定。"}
		return reply, true
	}
	f, err := client.Lookup(ctx, m.Latitude, m.Longitude, m.Address)
	if err != nil {
		text := "暫時無法取得天氣，請稍後再分享位置。"
		if errors.Is(err, weather.ErrOutsideCoverage) {
			text = "此位置不在目前支援的臺灣區域預報範圍內。"
		}
		reply.Message = &messaging_api.TextMessage{Text: text}
	} else {
		reply.Message = Weather(f)
	}
	return reply, true
}
