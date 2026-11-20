package main

import (
	"log/slog"
	"os"

	"github.com/963sup/line-mini-app-bot/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
