package attendance

import (
	"context"
	"time"
)

// LocatedStore binds a browser submission to its original signed group event.
type LocatedStore interface {
	Submit(context.Context, string, string, Position, time.Time) (Result, error)
	Groups(context.Context, string) ([]string, error)
	RecentGroup(context.Context, string, string) ([]Record, error)
}

type LaunchStore interface {
	CreateLaunch(context.Context, string, string) (string, error)
	LaunchGroup(context.Context, string) (string, error)
	ExchangeLaunch(context.Context, string, string) (string, error)
}

type Store interface {
	Punch(context.Context, string, string, string, string, time.Time, int64) (Result, error)
	Recent(context.Context, string) ([]Record, error)
}
