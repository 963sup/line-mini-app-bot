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

	httpadapter "github.com/963sup/line-mini-app-bot/internal/adapters/http"
	"github.com/963sup/line-mini-app-bot/internal/adapters/line"
)

func Run() (resultErr error) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	httpadapter.RegisterGeolocation(mux)
	appURL, err := httpadapter.RegisterMiniApp(mux)
	if err != nil {
		return err
	}
	attendance, closeAttendance, err := newAttendanceBot(context.Background(), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"))
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeAttendance()) }()
	if attendance != nil && os.Getenv("LINE_MINI_APP_LIFF_ID") != "" {
		attendance.AppURL = "https://liff.line.me/" + os.Getenv("LINE_MINI_APP_LIFF_ID")
	}
	var browserAttendance *httpadapter.Attendance
	if attendance != nil {
		browserAttendance = httpadapter.NewAttendance(attendance.Store, attendance.Zone, attendance.HistoryGroupName)
	}
	httpadapter.RegisterAttendanceAPI(mux, browserAttendance)
	summaries, closeSummaries, err := registerSummary(context.Background(), mux)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeSummaries()) }()
	if err := line.RegisterWebhook(mux, appURL, attendance, summaries); err != nil {
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
