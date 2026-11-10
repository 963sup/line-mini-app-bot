package attendance

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalidLeave  = errors.New("invalid leave request")
	ErrLeaveConflict = errors.New("leave retry payload conflict")
)

// Leave is a registration only. It never changes attendance or implies approval.
type Leave struct {
	ID        string
	GroupID   string
	Kind      string
	Start     time.Time
	End       time.Time
	Reason    string
	CreatedAt time.Time
}

type LeaveStore interface {
	CreateLeaveLaunch(context.Context, string) (string, error)
	LeaveGroup(context.Context, string) (string, error)
	SubmitLeave(context.Context, string, string, string, Leave) (Leave, error)
	RecentLeaves(context.Context, string) ([]Leave, error)
}

func NormalizeLeave(l Leave) (Leave, error) {
	l.Reason = strings.TrimSpace(l.Reason)
	l.Start, l.End = l.Start.UTC(), l.End.UTC()
	if l.Kind != "personal" && l.Kind != "sick" && l.Kind != "annual" && l.Kind != "other" {
		return Leave{}, ErrInvalidLeave
	}
	if l.Start.IsZero() || l.Start.Year() < 2000 || l.Start.Year() > 2100 || !l.End.After(l.Start) || l.End.Sub(l.Start) > 366*24*time.Hour || l.Reason == "" || utf8.RuneCountInString(l.Reason) > 500 {
		return Leave{}, ErrInvalidLeave
	}
	return l, nil
}

// ValidLeaveKey uses the same canonical UUID shape as other retry-safe submissions.
func ValidLeaveKey(key string) bool { return correctionKeyPattern.MatchString(key) }
