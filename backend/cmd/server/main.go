package main

import (
	"log/slog"
	"os"

	"github.com/963sup/line-mini-app-bot/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
