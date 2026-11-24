package entity

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("group not found")

// Group represents an engineering group within one Messaging API channel.
type Group struct {
	ID, GroupID, Name, Status string
	Version                   int64
	CreatedAt, UpdatedAt      time.Time
}
