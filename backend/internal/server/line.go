package server

import (
	"context"
	"net/http"
	"os"
	"time"

	attendance "github.com/963sup/line-mini-app-bot/internal/business/attendance/line"
	summary "github.com/963sup/line-mini-app-bot/internal/business/summary/line"
	"github.com/963sup/line-mini-app-bot/internal/business/weather/forecast/cwa"
	weather "github.com/963sup/line-mini-app-bot/internal/business/weather/forecast/line"
	navigation "github.com/963sup/line-mini-app-bot/internal/navigation/line"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func registerLINE(mux *http.ServeMux, appURL string, punches *attendance.Attendance, summaries *summary.Summary) (func(context.Context, string) string, error) {
	var client weather.ForecastLookup
	if key := os.Getenv("CWA_API_KEY"); key != "" {
		client = &cwa.Client{APIKey: key, HTTP: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	}
	lookup, err := transport.RegisterWebhook(mux, []transport.Handler{
		{Handle: summaries.Observe, Failure: "summary temporarily unavailable"},
		{Handle: punches.Reply, Failure: "attendance unavailable; retry safely", RetryOnSendFailure: true},
		{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
			reply, handled := weather.ReplyForEvent(ctx, event, client)
			return reply, handled, nil
		}},
		{Handle: func(_ context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
			return navigation.ReplyForEvent(event, appURL), true, nil
		}},
	})
	return lookup, err
}
