package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	attendance "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/line"
	membershipline "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/line"
	membershipapp "github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	objectline "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/line"
	organizationline "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/line"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/line"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	summary "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/line"
	groupline "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/line"
	"github.com/963sup/line-mini-app-bot/internal/business/weather/adapter/cwa/forecast"
	weather "github.com/963sup/line-mini-app-bot/internal/business/weather/adapter/line/forecast"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	navigation "github.com/963sup/line-mini-app-bot/internal/transport/line/navigation"
)

func registerLINE(mux *http.ServeMux, config messagingConfig, bot *lineplatform.Messaging, appURL, endpointURL string, punches *attendance.Attendance, summaries *summary.Summary, repositories *repositoryline.Observer, groups *groupline.Observer, catalog *repositoryApplication.Service, creation organizationline.Conversation, invitations *membershipapp.Service) func(context.Context, string) string {
	if bot == nil {
		slog.Info("LINE webhook disabled: Messaging API configuration is absent")
		return nil
	}
	lookup := bot.GroupName
	var client weather.ForecastLookup
	if key := os.Getenv("CWA_API_KEY"); key != "" {
		client = &cwa.Client{APIKey: key, HTTP: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	}
	var repositoryID func(context.Context, string) (string, error)
	if repositories != nil {
		repositoryID = func(ctx context.Context, group string) (string, error) {
			repo, err := repositories.Bindings.Repository(ctx, group)
			if errors.Is(err, repositoryEntity.ErrNotFound) {
				return "", nil
			}
			return repo.ID, err
		}
	}
	dispatch := transport.Dispatch{
		EarlyCommands: []transport.Handler{
			membershipline.Handler(invitations),
			organizationline.CreationHandler(creation, appURL),
			repositoryline.EngineeringHandler(appURL, catalog),
		},
		Observers: []transport.Observer{
			{Observe: groups.Observe, Failure: "space temporarily unavailable"},
			{Observe: repositories.Observe, Failure: "repository temporarily unavailable"},
		},
		Commands: []transport.Handler{
			summaries.Handler(),
			{Handle: summaries.Collect, Failure: "summary temporarily unavailable"},
			punches.Handler(),
			weather.Handler(client),
			objectline.Handler(appURL),
			navigation.Menu{StickerPackageID: os.Getenv("LINE_MENU_STICKER_PACKAGE_ID"), StickerID: os.Getenv("LINE_MENU_STICKER_ID"), PrivateActions: attendance.NavigationActions()}.Handler(appURL, endpointURL, repositoryID),
		},
	}
	if repositories != nil {
		repositories.GroupName = lookup
	}
	if groups != nil {
		groups.GroupName = lookup
	}
	transport.RegisterWebhook(mux, config.channel, config.secret, bot.Reply, dispatch)
	return lookup
}
