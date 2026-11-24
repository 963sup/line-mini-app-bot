package bootstrap

import (
	"context"
	"log/slog"
	"time"
)

// Both workers reuse their durable checkpoints after a revision restarts.
func runDelivery(ctx context.Context, operation string, timeout time.Duration, drain func(context.Context) error) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batch, cancel := context.WithTimeout(ctx, timeout)
			err := drain(batch)
			cancel()
			if err != nil && ctx.Err() == nil {
				slog.WarnContext(ctx, "background delivery deferred", "operation", operation)
			}
		}
	}
}
