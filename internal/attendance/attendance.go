// Package attendance defines the group-scoped attendance contract.
package attendance

import (
	"context"
	"time"
)

type Record struct {
	SchemaVersion int       `firestore:"schemaVersion"`
	GroupID       string    `firestore:"groupID"`
	ClockInAt     time.Time `firestore:"clockInAt"`
	ClockOutAt    time.Time `firestore:"clockOutAt,omitempty"`
}

type Result struct {
	Code   string `firestore:"code"`
	Record Record `firestore:"record"`
}

type Store interface {
	Punch(context.Context, string, string, string, string, time.Time, int64) (Result, error)
	Recent(context.Context, string) ([]Record, error)
}
