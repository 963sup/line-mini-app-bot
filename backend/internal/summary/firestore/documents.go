package firestore

import (
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/summary"
)

type messageDocument struct {
	ID, EventID, UserID, Text string
	OccurredAt, ExpireAt      time.Time
	Epoch                     int64
	Deleted                   bool
	EditedAt                  int64
}

func decodeMessage(snap *db.DocumentSnapshot, target *summary.Message) error {
	var value messageDocument
	if err := snap.DataTo(&value); err != nil {
		return err
	}
	*target = summary.Message(value)
	return nil
}

type jobDocument struct {
	ID, GroupID, Kind, Result, RetryKey, Status string
	Epoch, Revision                             int64
	Cutoff, ExpireAt, LeaseUntil                time.Time
	Lease                                       string
	// Sources contain identifiers only; no additional copy of chat content.
	Sources []string
}

func decodeJob(snap *db.DocumentSnapshot, target *summary.Job) error {
	var value jobDocument
	if err := snap.DataTo(&value); err != nil {
		return err
	}
	*target = summary.Job(value)
	return nil
}
