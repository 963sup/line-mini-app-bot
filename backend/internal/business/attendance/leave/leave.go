package leave

import (
	"context"
	"errors"
	"regexp"
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
	CreateLeaveLaunch(ctx context.Context, groupID string) (string, error)
	LeaveGroup(ctx context.Context, token string) (string, error)
	SubmitLeave(ctx context.Context, token, userID, retryKey string, proposal Leave) (Leave, error)
	RecentLeaves(ctx context.Context, userID string) ([]Leave, error)
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
func ValidLeaveKey(key string) bool { return leaveKeyPattern.MatchString(key) }

var leaveKeyPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
