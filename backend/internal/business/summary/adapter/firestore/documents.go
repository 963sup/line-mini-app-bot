package firestore

import (
	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

import (
	"time"

	db "cloud.google.com/go/firestore"
)

type messageDocument struct {
	ID, EventID, UserID, Text string
	OccurredAt, ExpireAt      time.Time
	Epoch                     int64
	Deleted                   bool
	EditedAt                  int64
}

func decodeMessage(snap *db.DocumentSnapshot, target *summaryDomainEntityMessagecollection.Message) error {
	var value messageDocument
	if err := snap.DataTo(&value); err != nil {
		return err
	}
	*target = summaryDomainEntityMessagecollection.Message(value)
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

func decodeJob(snap *db.DocumentSnapshot, target *summaryDomainEntitySummarization.Job) error {
	var value jobDocument
	if err := snap.DataTo(&value); err != nil {
		return err
	}
	*target = summaryDomainEntitySummarization.Job(value)
	return nil
}
