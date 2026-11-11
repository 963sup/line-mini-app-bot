package firestore

import (
	"time"

	db "cloud.google.com/go/firestore"
	correctionsmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/corrections"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/punch"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/worktime"
)

type positionDocument struct {
	Latitude       float64   `json:"latitude" firestore:"latitude"`
	Longitude      float64   `json:"longitude" firestore:"longitude"`
	AccuracyMeters float64   `json:"accuracyMeters" firestore:"accuracyMeters"`
	CapturedAt     time.Time `json:"capturedAt" firestore:"capturedAt"`
}
type recordDocument struct {
	IsOpen           bool                `firestore:"isOpen" json:"-"`
	Revision         int64               `firestore:"revision"`
	CorrectionID     string              `firestore:"correctionID,omitempty"`
	RebuildIssues    []string            `firestore:"-"`
	LastAppliedAt    time.Time           `firestore:"lastAppliedAt"`
	SourceEventIDs   []string            `firestore:"sourceEventIDs,omitempty"`
	EventSequence    int64               `firestore:"eventSequence"`
	RecordID         string              `firestore:"recordID,omitempty"`
	Breaks           []worktime.Interval `firestore:"breaks,omitempty"`
	BreakStartedAt   time.Time           `firestore:"breakStartedAt,omitempty"`
	RestConfirmed    bool                `firestore:"restConfirmed"`
	SchemaVersion    int                 `firestore:"schemaVersion"`
	GroupID          string              `firestore:"groupID"`
	ClockInAt        time.Time           `firestore:"clockInAt"`
	ClockOutAt       time.Time           `firestore:"clockOutAt,omitempty"`
	ClockInPosition  *positionDocument   `firestore:"clockInPosition,omitempty"`
	ClockOutPosition *positionDocument   `firestore:"clockOutPosition,omitempty"`
}
type requestDocument struct {
	GroupID   string    `firestore:"groupID"`
	UserID    string    `firestore:"userKey"`
	EventID   string    `firestore:"eventID"`
	Action    string    `firestore:"action"`
	EventTime int64     `firestore:"eventTime"`
	ExpiresAt time.Time `firestore:"expiresAt"`
}
type resultDocument struct {
	Code   string         `firestore:"code"`
	Record recordDocument `firestore:"record"`
}
type correctionDocument struct {
	ApplicantName         string              `firestore:"applicantName,omitempty" json:"-"`
	NewRecord             bool                `firestore:"newRecord,omitempty" json:"newRecord,omitempty"`
	ID                    string              `firestore:"id" json:"id"`
	UserKey               string              `firestore:"userKey" json:"-"`
	GroupID               string              `firestore:"groupID" json:"groupId"`
	RecordID              string              `firestore:"recordID" json:"recordId"`
	ExpectedEventSequence int64               `firestore:"expectedEventSequence" json:"expectedEventSequence"`
	ExpectedRevision      int64               `firestore:"expectedRevision" json:"expectedRevision"`
	Reason                string              `firestore:"reason" json:"reason"`
	Start                 time.Time           `firestore:"start" json:"start"`
	End                   time.Time           `firestore:"end" json:"end"`
	Breaks                []worktime.Interval `firestore:"breaks" json:"breaks"`
	RestConfirmed         bool                `firestore:"restConfirmed" json:"restConfirmed"`
	CreatedAt             time.Time           `firestore:"createdAt" json:"-"`
}

func positionDocumentFrom(v *recordmodel.Position) *positionDocument {
	if v == nil {
		return nil
	}
	return &positionDocument{Latitude: v.Latitude, Longitude: v.Longitude, AccuracyMeters: v.AccuracyMeters, CapturedAt: v.CapturedAt}
}

func positionDomain(v *positionDocument) *recordmodel.Position {
	if v == nil {
		return nil
	}
	return &recordmodel.Position{Latitude: v.Latitude, Longitude: v.Longitude, AccuracyMeters: v.AccuracyMeters, CapturedAt: v.CapturedAt}
}

func recordDocumentFrom(v recordmodel.Record) recordDocument {
	return recordDocument{IsOpen: v.IsOpen, Revision: v.Revision, CorrectionID: v.CorrectionID, RebuildIssues: v.RebuildIssues, LastAppliedAt: v.LastAppliedAt, SourceEventIDs: v.SourceEventIDs, EventSequence: v.EventSequence, RecordID: v.RecordID, Breaks: v.Breaks, BreakStartedAt: v.BreakStartedAt, RestConfirmed: v.RestConfirmed, SchemaVersion: v.SchemaVersion, GroupID: v.GroupID, ClockInAt: v.ClockInAt, ClockOutAt: v.ClockOutAt, ClockInPosition: positionDocumentFrom(v.ClockInPosition), ClockOutPosition: positionDocumentFrom(v.ClockOutPosition)}
}

func (v recordDocument) domain() recordmodel.Record {
	return recordmodel.Record{IsOpen: v.IsOpen, Revision: v.Revision, CorrectionID: v.CorrectionID, RebuildIssues: v.RebuildIssues, LastAppliedAt: v.LastAppliedAt, SourceEventIDs: v.SourceEventIDs, EventSequence: v.EventSequence, RecordID: v.RecordID, Breaks: v.Breaks, BreakStartedAt: v.BreakStartedAt, RestConfirmed: v.RestConfirmed, SchemaVersion: v.SchemaVersion, GroupID: v.GroupID, ClockInAt: v.ClockInAt, ClockOutAt: v.ClockOutAt, ClockInPosition: positionDomain(v.ClockInPosition), ClockOutPosition: positionDomain(v.ClockOutPosition)}
}

func decodeRecord(snapshot *db.DocumentSnapshot, target *recordmodel.Record) error {
	var value recordDocument
	if err := snapshot.DataTo(&value); err != nil {
		return err
	}
	*target = value.domain()
	return nil
}

func resultDocumentFrom(v punch.Result) resultDocument {
	return resultDocument{Code: v.Code, Record: recordDocumentFrom(v.Record)}
}

func (v resultDocument) domain() punch.Result {
	return punch.Result{Code: v.Code, Record: v.Record.domain()}
}

func decodeResult(snapshot *db.DocumentSnapshot, target *punch.Result) error {
	var value resultDocument
	if err := snapshot.DataTo(&value); err != nil {
		return err
	}
	*target = value.domain()
	return nil
}

func correctionDocumentFrom(v correctionsmodel.Correction) correctionDocument {
	return correctionDocument{ApplicantName: v.ApplicantName, NewRecord: v.NewRecord, ID: v.ID, UserKey: v.UserKey, GroupID: v.GroupID, RecordID: v.RecordID, ExpectedEventSequence: v.ExpectedEventSequence, ExpectedRevision: v.ExpectedRevision, Reason: v.Reason, Start: v.Start, End: v.End, Breaks: v.Breaks, RestConfirmed: v.RestConfirmed, CreatedAt: v.CreatedAt}
}

func (v correctionDocument) domain() correctionsmodel.Correction {
	return correctionsmodel.Correction{ApplicantName: v.ApplicantName, NewRecord: v.NewRecord, ID: v.ID, UserKey: v.UserKey, GroupID: v.GroupID, RecordID: v.RecordID, ExpectedEventSequence: v.ExpectedEventSequence, ExpectedRevision: v.ExpectedRevision, Reason: v.Reason, Start: v.Start, End: v.End, Breaks: v.Breaks, RestConfirmed: v.RestConfirmed, CreatedAt: v.CreatedAt}
}

func decodeCorrection(snapshot *db.DocumentSnapshot, target *correctionsmodel.Correction) error {
	var value correctionDocument
	if err := snapshot.DataTo(&value); err != nil {
		return err
	}
	*target = value.domain()
	return nil
}
