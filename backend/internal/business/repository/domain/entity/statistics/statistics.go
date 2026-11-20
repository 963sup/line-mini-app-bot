package statistics

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("repository not found")

var ErrUnavailable = errors.New("statistics unavailable")

type Statistics struct {
	OpenIssues, InProgressIssues, CompletedIssues, Discussions, Milestones, Projects int64
	ObservedAt                                                                       time.Time
}
