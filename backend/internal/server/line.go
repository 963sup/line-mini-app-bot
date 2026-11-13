package server

import (
	"context"
	"net/http"
	"os"
	"time"

	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/line"
	spaceline "github.com/963sup/line-mini-app-bot/internal/business/space/line"

	attendance "github.com/963sup/line-mini-app-bot/internal/business/attendance/line"
	summary "github.com/963sup/line-mini-app-bot/internal/business/summary/line"
	"github.com/963sup/line-mini-app-bot/internal/business/weather/forecast/cwa"
	weather "github.com/963sup/line-mini-app-bot/internal/business/weather/forecast/line"
	navigation "github.com/963sup/line-mini-app-bot/internal/navigation/line"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func registerLINE(mux *http.ServeMux, appURL string, punches *attendance.Attendance, summaries *summary.Summary, repositories *repositoryline.Observer, spaces *spaceline.Observer) (func(context.Context, string) string, error) {
	var client weather.ForecastLookup
	if key := os.Getenv("CWA_API_KEY"); key != "" {
		client = &cwa.Client{APIKey: key, HTTP: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	}
	lookup, err := transport.RegisterWebhook(mux, []transport.Handler{
		{Handle: spaces.Observe, Failure: "space temporarily unavailable"},
		{Handle: repositories.Observe, Failure: "repository temporarily unavailable"},
		{Handle: summaries.Observe, Failure: "summary temporarily unavailable"},
		{Handle: punches.Reply, Failure: "attendance unavailable; retry safely", RetryOnSendFailure: true},
		{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
			reply, handled := weather.ReplyForEvent(ctx, event, client)
			return reply, handled, nil
		}},
		{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
			reply := navigation.ReplyForEvent(event, appURL)
			if reply.Message == nil || repositories == nil {
				return reply, true, nil
			}
			var source webhook.SourceInterface
			switch e := event.(type) {
			case webhook.JoinEvent:
				source = e.Source
			case webhook.MessageEvent:
				source = e.Source
			case webhook.PostbackEvent:
				source = e.Source
			}
			if group, ok := source.(webhook.GroupSource); ok {
				container, err := repositories.Store.GetByGroupID(ctx, group.GroupId)
				if err != nil {
					return transport.Reply{}, false, err
				}
				reply = navigation.ReplyForEvent(event, appURL, container.ID)
			}
			return reply, true, nil
		}},
	})
	if repositories != nil {
		repositories.GroupName = lookup
	}
	if spaces != nil {
		spaces.GroupName = lookup
	}
	return lookup, err
}
