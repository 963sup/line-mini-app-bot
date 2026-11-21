package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/963sup/line-mini-app-bot/internal/bootstrap"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "sync-line-navigation":
			if len(os.Args) != 2 {
				fmt.Fprintln(os.Stderr, "sync-line-navigation takes no arguments")
				os.Exit(2)
			}
			if err := bootstrap.SyncLINENavigation(context.Background()); err != nil {
				slog.Error("LINE navigation publication failed", "error", err)
				os.Exit(1)
			}
			slog.Info("LINE navigation publication verified")
		case "check-firestore":
			if err := bootstrap.CheckFirestore(context.Background(), os.Args[2:]); err != nil {
				slog.Error("Firestore check failed", "error", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintln(os.Stderr, "unsupported application mode")
			os.Exit(2)
		}
		return
	}
	if err := bootstrap.Run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
