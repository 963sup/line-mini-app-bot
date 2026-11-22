package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	attendance "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/line"
	membershipline "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/line"
	membershipapp "github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	organizationline "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/line"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/line"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	summary "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/line"
	"github.com/963sup/line-mini-app-bot/internal/business/weather/adapter/cwa/forecast"
	weather "github.com/963sup/line-mini-app-bot/internal/business/weather/adapter/line/forecast"
	groupline "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/line"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	navigation "github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func registerLINE(mux *http.ServeMux, appURL, endpointURL string, punches *attendance.Attendance, summaries *summary.Summary, repositories *repositoryline.Observer, groups *groupline.Observer, catalog *repositoryApplication.Service, creation organizationline.Conversation, invitations *membershipapp.Service) (func(context.Context, string) string, error) {
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
		return nil, err
	}
	lookup := lineplatform.GroupNameLookup(bot)
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
	dispatch := transport.Dispatch{
		EarlyCommands: []transport.Handler{
			membershipline.Handler(invitations),
			organizationline.CreationHandler(creation, appURL),
			{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
				return repositoryline.MyEngineering(ctx, event, appURL, catalog)
			}, Failure: "repository temporarily unavailable"},
		},
		Observers: []transport.Observer{
			{Observe: groups.Observe, Failure: "space temporarily unavailable"},
			{Observe: repositories.Observe, Failure: "repository temporarily unavailable"},
		},
		Commands: []transport.Handler{
			{Handle: summaries.Reply, Failure: "summary temporarily unavailable"},
			{Handle: summaries.Collect, Failure: "summary temporarily unavailable"},
			{Handle: punches.Reply, Failure: "attendance unavailable; retry safely", RetryOnSendFailure: true},
			{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
				reply, handled := weather.ReplyForEvent(ctx, event, client)
				return reply, handled, nil
			}},
			navigation.Handler(appURL, endpointURL, repositoryID),
		},
	}
	if repositories != nil {
		repositories.GroupName = lookup
	}
	if groups != nil {
		groups.GroupName = lookup
	}
	path := "/webhooks/line/" + channelID
	mux.HandleFunc("POST "+path, transport.WebhookHandler(secret, bot, dispatch))
	slog.Info("LINE webhook enabled", "path", path)
	return lookup, nil
}
