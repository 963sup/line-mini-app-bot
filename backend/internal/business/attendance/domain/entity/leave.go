package entity

import (
	"errors"
	"time"
)

var ErrInvalidLeave = errors.New("invalid leave request")

var ErrLeaveConflict = errors.New("leave retry payload conflict")

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
