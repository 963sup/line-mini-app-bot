package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/browser"
	attendancehttp "github.com/963sup/line-mini-app-bot/internal/business/attendance/http"
	attendanceline "github.com/963sup/line-mini-app-bot/internal/business/attendance/line"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/presentation"
	discussionhttp "github.com/963sup/line-mini-app-bot/internal/business/discussion/http"
	issuehttp "github.com/963sup/line-mini-app-bot/internal/business/issue/http"
	objecthttp "github.com/963sup/line-mini-app-bot/internal/business/object/http"
	objectimporthttp "github.com/963sup/line-mini-app-bot/internal/business/objectimport/http"
	projecthttp "github.com/963sup/line-mini-app-bot/internal/business/project/http"
	repositoryhttp "github.com/963sup/line-mini-app-bot/internal/business/repository/http"
	labelhttp "github.com/963sup/line-mini-app-bot/internal/business/repository/labels/http"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/line"
	milestonehttp "github.com/963sup/line-mini-app-bot/internal/business/repository/milestones/http"
	statisticshttp "github.com/963sup/line-mini-app-bot/internal/business/repository/statistics/http"
	spacehttp "github.com/963sup/line-mini-app-bot/internal/business/space/http"
	spaceline "github.com/963sup/line-mini-app-bot/internal/business/space/line"
	"github.com/963sup/line-mini-app-bot/internal/diagnostics/geolocation"
	identityhttp "github.com/963sup/line-mini-app-bot/internal/identity/http"
	identityline "github.com/963sup/line-mini-app-bot/internal/identity/line"
)

func Run() (resultErr error) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	geolocation.RegisterGeolocation(mux)
	if err := browser.Register(mux); err != nil {
		return err
	}
	auth := &identityhttp.Authenticator{ChannelID: os.Getenv("LINE_MINI_APP_CHANNEL_ID"), LIFFID: os.Getenv("LINE_MINI_APP_LIFF_ID")}
	auth.Verifier = &identityline.Verifier{ChannelID: auth.ChannelID, Client: &http.Client{Timeout: 5 * time.Second}}
	sessions, closeSessions, err := newSessions(context.Background(), auth.ChannelID)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeSessions()) }()
	auth.Sessions = sessions
	appURL, err := auth.Register(mux)
	if err != nil {
		return err
	}
	attendance, closeAttendance, err := newAttendance(context.Background(), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"))
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeAttendance()) }()
	var groupLookup func(context.Context, string) string
	lookup := func(ctx context.Context, group string) string {
		if groupLookup != nil {
			return groupLookup(ctx, group)
		}
		return ""
	}
	var punches *attendanceline.Attendance
	var browserAttendance *attendancehttp.Attendance
	if attendance != nil {
		punches = &attendanceline.Attendance{Store: attendance.store, Zone: attendance.zone, GroupName: lookup}
		if id := os.Getenv("LINE_MINI_APP_LIFF_ID"); id != "" {
			punches.AppURL = "https://liff.line.me/" + id
		}
		browserAttendance = attendancehttp.NewAttendance(attendance.store, attendance.zone, func(ctx context.Context, group string) string {
			return presentation.GroupName(group, lookup(ctx, group))
		})
	}

	attendancehttp.RegisterAttendanceAPI(mux, browserAttendance, auth)
	summaries, closeSummaries, err := registerSummary(context.Background(), mux)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeSummaries()) }()
	repositories, closeRepositories, err := newRepositories(context.Background())
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeRepositories()) }()
	spaces, closeSpaces, err := newSpaces(context.Background())
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeSpaces()) }()
	spacehttp.Register(mux, spaces, auth)
	repositoryhttp.Register(mux, repositories, auth)
	labelhttp.Register(mux, newLabels(repositories), auth)
	milestonehttp.Register(mux, newMilestones(repositories), auth)
	projecthttp.Register(mux, newProjects(spaces, repositories), auth)
	statisticshttp.Register(mux, newStatistics(repositories), auth)
	issuehttp.Register(mux, newIssues(repositories), auth)
	objecthttp.Register(mux, newObjects(repositories), auth)
	objectimporthttp.Register(mux, newObjectImports(repositories), auth, newObjectParser(repositories))
	discussionhttp.Register(mux, newDiscussions(repositories), auth)
	var repositoryObserver *repositoryline.Observer
	if repositories != nil {
		repositoryObserver = &repositoryline.Observer{Store: repositories.Store}
	}
	var spaceObserver *spaceline.Observer
	if spaces != nil {
		spaceObserver = &spaceline.Observer{Store: spaces.Store}
	}
	groupLookup, err = registerLINE(mux, appURL, punches, summaries, repositoryObserver, spaceObserver)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr: net.JoinHostPort("", port), Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 100 * time.Second, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	result := make(chan error, 1)
	slog.Info("server starting", "port", port)
	go func() { result <- server.ListenAndServe() }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
