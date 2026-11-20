package bootstrap

import (
	"context"
	"net/http"
	"os"
	"time"

	attendance "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/line"
	groupline "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/line"
	organizationline "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/line"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/line"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	summary "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/line"
	"github.com/963sup/line-mini-app-bot/internal/business/weather/adapter/cwa/forecast"
	weather "github.com/963sup/line-mini-app-bot/internal/business/weather/adapter/line/forecast"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	navigation "github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func registerLINE(mux *http.ServeMux, appURL, endpointURL string, punches *attendance.Attendance, summaries *summary.Summary, repositories *repositoryline.Observer, groups *groupline.Observer, catalog *repositoryApplication.Service, creation organizationline.Conversation) (func(context.Context, string) string, error) {
	var client weather.ForecastLookup
	if key := os.Getenv("CWA_API_KEY"); key != "" {
		client = &cwa.Client{APIKey: key, HTTP: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	}
	var repositoryID func(context.Context, string) (string, error)
	if repositories != nil {
		repositoryID = func(ctx context.Context, group string) (string, error) {
			repo, err := repositories.Store.GetByGroupID(ctx, group)
			return repo.ID, err
		}
	}
	lookup, err := transport.RegisterWebhook(mux, []transport.Handler{
		organizationline.CreationHandler(creation, appURL),
		{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
			return repositoryline.MyEngineering(ctx, event, appURL, catalog)
		}, Failure: "repository temporarily unavailable"},
		{Handle: groups.Observe, Failure: "space temporarily unavailable"},
		{Handle: repositories.Observe, Failure: "repository temporarily unavailable"},
		{Handle: summaries.Observe, Failure: "summary temporarily unavailable"},
		{Handle: punches.Reply, Failure: "attendance unavailable; retry safely", RetryOnSendFailure: true},
		{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
			reply, handled := weather.ReplyForEvent(ctx, event, client)
			return reply, handled, nil
		}},
		navigation.Handler(appURL, endpointURL, repositoryID),
	})
	if repositories != nil {
		repositories.GroupName = lookup
	}
	if groups != nil {
		groups.GroupName = lookup
	}
	return lookup, err
}
