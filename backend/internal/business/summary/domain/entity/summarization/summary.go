package summarization

import "time"

type Job struct {
	ID, GroupID, Kind, Result, RetryKey, Status string
	Epoch, Revision                             int64
	Cutoff, ExpireAt, LeaseUntil                time.Time
	Lease                                       string
	// Sources contain identifiers only; no additional copy of chat content.
	Sources []string
}
