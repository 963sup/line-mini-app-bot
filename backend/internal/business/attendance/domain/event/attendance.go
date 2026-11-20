package event

import (
	"time"
)

type Event struct {
	ContractVersion   int
	RecordID, GroupID string
	ID                string
	Kind              string
	At                time.Time
	Source            string
	Sequence          int64
	ResultCode        string
}
