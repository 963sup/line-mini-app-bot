package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Run drains one bounded batch per tick and stops when the server shuts down.
func Run(ctx context.Context, drain func(context.Context) error) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batch, cancel := context.WithTimeout(ctx, 50*time.Second)
			err := drain(batch)
			cancel()
			if err != nil && ctx.Err() == nil {
				slog.Warn("object quantity event delivery deferred")
			}
		}
	}
}
