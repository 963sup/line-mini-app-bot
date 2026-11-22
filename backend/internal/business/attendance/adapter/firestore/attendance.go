package firestore

import (
	"context"
	"fmt"
	"time"

	db "cloud.google.com/go/firestore"
	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	attendanceevent "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/event"
	attendanceservice "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/service"
	attendancevalueobject "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/valueobject"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Attendance struct {
	Client    *db.Client
	ChannelID string
}

func (s *Attendance) user(user string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("users").Doc(firestorekey.Encode(user))
}

type punchMode uint8

const (
	recordPunch punchMode = iota
	resolvePunch
)

type punchCommand struct {
	GroupID, UserID, EventID, Action string
	ReceivedAt                       time.Time
	EventTimeMillis                  int64
	Position                         *attendancevalueobject.Position
	RequestID                        string
	Mode                             punchMode
}

// executePunch atomically resolves the receipt, transition and persistence.
// Mode is explicit: a missing position must never implicitly select preflight.
func (s *Attendance) executePunch(ctx context.Context, command punchCommand) (attendancedto.PunchResult, error) {
	u := s.user(command.UserID)
	state := u.Collection("attendanceStates").Doc(firestorekey.Encode(command.GroupID))
	receipt := u.Collection("attendanceEvents").Doc(firestorekey.Encode(command.GroupID, command.EventID, command.Action))
	var result attendancedto.PunchResult
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = attendancedto.PunchResult{}
		var request requestDocument
		if command.RequestID != "" {
			snap, err := tx.Get(s.requests().Doc(command.RequestID))
			if err != nil {
				return err
			}
			if err := snap.DataTo(&request); err != nil {
				return err
			}
			if request.UserID != firestorekey.Encode(command.UserID) || request.GroupID != command.GroupID || request.EventID != command.EventID || request.Action != command.Action {
				return attendanceport.ErrInvalidRequest
			}
		}
		old, err := tx.Get(receipt)
		if err == nil {
			return decodeResult(old, &result)
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		if command.RequestID != "" && !command.ReceivedAt.Before(request.ExpiresAt) {
			return attendanceport.ErrInvalidRequest
		}
		var active struct {
			RecordID      string `firestore:"recordID"`
			LastEventTime int64  `firestore:"lastEventTime"`
		}
		snapshot, err := tx.Get(state)
		if err == nil {
			if err := snapshot.DataTo(&active); err != nil {
				return err
			}
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		if command.EventTimeMillis < active.LastEventTime {
			result.Code = "stale_event"
			if err := tx.Create(u.Collection("attendanceRawEvents").Doc(firestorekey.Encode(command.GroupID, command.EventID, command.Action)), map[string]any{"eventID": command.EventID, "action": command.Action, "occurredAt": command.ReceivedAt, "eventTime": command.EventTimeMillis, "groupID": command.GroupID, "source": "server", "resultCode": result.Code}); err != nil {
				return err
			}
			return tx.Create(receipt, resultDocumentFrom(result))
		}
		var record attendanceentity.Record
		var ref *db.DocumentRef
		if active.RecordID != "" {
			ref = u.Collection("attendance").Doc(active.RecordID)
			snapshot, err := tx.Get(ref)
			if err != nil {
				return err
			}
			if err := decodeRecord(snapshot, &record); err != nil {
				return err
			}
			if record.GroupID != command.GroupID || !record.ClockOutAt.IsZero() {
				return fmt.Errorf("invalid attendance state")
			}
		}
		machine := attendanceentity.State{}
		if active.RecordID != "" {
			machine.Active = &attendanceentity.Session{ID: active.RecordID, Start: record.ClockInAt, End: record.ClockOutAt, Breaks: record.Breaks, RestConfirmed: record.RestConfirmed, EventIDs: record.SourceEventIDs}
			machine.RestStartedAt = record.BreakStartedAt
			machine.LastAt = record.LastAppliedAt
			if machine.LastAt.IsZero() {
				machine.LastAt = record.ClockInAt
			}
			for _, rest := range record.Breaks {
				if rest.End.After(machine.LastAt) {
					machine.LastAt = rest.End
				}
			}
			if record.BreakStartedAt.After(machine.LastAt) {
				machine.LastAt = record.BreakStartedAt
			}
		}
		next, outcome := attendanceservice.ApplyEvent(machine, attendanceevent.Event{ID: command.EventID, Kind: command.Action, At: command.ReceivedAt})
		result.Code = outcome.Code
		// A preflight shares the authoritative transition and transaction reads.
		// It can seal a rejected outcome, but must never accept a located event
		// or consume its receipt before a position has been submitted.
		if outcome.Accepted && command.Mode == resolvePunch {
			result.Code = "needs_position"
			return nil
		}
		if outcome.Accepted {
			if command.Action == "clock-in" {
				record = attendanceentity.Record{SchemaVersion: 3, GroupID: command.GroupID, RecordID: firestorekey.Encode(command.GroupID, command.EventID), ClockInPosition: command.Position}
				ref = u.Collection("attendance").Doc(record.RecordID)
			}
			record.ClockInAt = outcome.Session.Start
			record.ClockOutAt = outcome.Session.End
			record.IsOpen = record.ClockOutAt.IsZero()
			record.Breaks = outcome.Session.Breaks
			record.RestConfirmed = outcome.Session.RestConfirmed
			record.BreakStartedAt = next.RestStartedAt
			record.EventSequence++
			record.LastAppliedAt = command.ReceivedAt
			record.SourceEventIDs = outcome.Session.EventIDs
			if command.Action == "clock-out" {
				record.ClockOutPosition = command.Position
			}
			if command.Action == "clock-in" {
				if err := tx.Create(ref, recordDocumentFrom(record)); err != nil {
					return err
				}
			} else {
				if err := tx.Set(ref, recordDocumentFrom(record)); err != nil {
					return err
				}
			}
			if err := tx.Set(u.Collection("attendanceGroups").Doc(firestorekey.Encode(command.GroupID)), map[string]any{"groupID": command.GroupID}); err != nil {
				return err
			}
			active.RecordID = ""
			if next.Active != nil {
				active.RecordID = record.RecordID
			}
		}
		result.Record = record

		if err := tx.Set(state, map[string]any{"recordID": active.RecordID, "lastEventTime": command.EventTimeMillis}); err != nil {
			return err
		}
		// A separate immutable event retains the attempted action and its outcome.
		raw := map[string]any{"eventID": command.EventID, "action": command.Action, "occurredAt": command.ReceivedAt, "source": "server", "eventTime": command.EventTimeMillis, "groupID": command.GroupID, "recordID": active.RecordID, "resultCode": result.Code, "sequence": record.EventSequence, "contractVersion": 1}
		if command.Position != nil {
			raw["position"] = positionDocumentFrom(command.Position)
			raw["source"] = "located-web"
		}
		if result.Record.RecordID != "" {
			raw["recordID"] = result.Record.RecordID
		}
		if err := tx.Create(u.Collection("attendanceRawEvents").Doc(firestorekey.Encode(command.GroupID, command.EventID, command.Action)), raw); err != nil {
			return err
		}
		return tx.Create(receipt, resultDocumentFrom(result))
	})
	return result, err
}

func (s *Attendance) requests() *db.CollectionRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("punchRequests")
}

func (s *Attendance) executePunchRequest(ctx context.Context, command punchCommand) (attendancedto.PunchResult, error) {
	snapshot, err := s.requests().Doc(command.RequestID).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return attendancedto.PunchResult{}, attendanceport.ErrInvalidRequest
	}
	if err != nil {
		return attendancedto.PunchResult{}, err
	}
	var request requestDocument
	if err := snapshot.DataTo(&request); err != nil {
		return attendancedto.PunchResult{}, err
	}
	if request.UserID != firestorekey.Encode(command.UserID) {
		return attendancedto.PunchResult{}, attendanceport.ErrInvalidRequest
	}
	command.GroupID, command.EventID, command.Action = request.GroupID, request.EventID, request.Action
	command.EventTimeMillis = request.EventTime
	return s.executePunch(ctx, command)
}

func (s *Attendance) Groups(ctx context.Context, user string) (_ []string, err error) {
	defer func() { err = translateError(err, nil) }()
	groups := []string{}
	cursor := ""
	for {
		page, err := s.GroupsPage(ctx, user, cursor, 100)
		if err != nil {
			return nil, err
		}
		groups = append(groups, page.Items...)
		if page.NextCursor == "" {
			return groups, nil
		}
		cursor = page.NextCursor
	}
}

func (s *Attendance) RecentGroup(ctx context.Context, user, group string) (_ []attendanceentity.Record, err error) {
	defer func() { err = translateError(err, nil) }()
	docs, err := s.user(user).Collection("attendance").Where("groupID", "==", group).OrderBy("clockInAt", db.Desc).OrderBy(db.DocumentID, db.Desc).Limit(5).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	return decodeAttendance(docs)
}

// RecordsOverlappingMonth includes every shift overlapping the reporting period, including open shifts.
func (s *Attendance) RecordsOverlappingMonth(ctx context.Context, user, group string, start time.Time) (_ []attendanceentity.Record, err error) {
	defer func() { err = translateError(err, nil) }()
	records, err := s.monthCandidates(ctx, user, group, start)
	if err != nil {
		return nil, err
	}
	records, err = s.verifyAndRebuildRecords(ctx, user, group, records)
	if err != nil {
		return nil, err
	}
	result := make([]attendanceentity.Record, 0, len(records))
	for _, r := range records {
		if overlapsMonth(r, start) {
			result = append(result, r)
		}
	}
	return result, nil
}

// verifyAndRebuildRecords preserves raw-event and correction evidence checks for every read shape.
func (s *Attendance) verifyAndRebuildRecords(ctx context.Context, user, group string, records []attendanceentity.Record) ([]attendanceentity.Record, error) {
	rawDocs := []*db.DocumentSnapshot{}
	for _, r := range records {
		if r.SchemaVersion < 3 || r.Revision > 0 {
			continue
		}
		docs, e := readQuery(ctx, s.user(user).Collection("attendanceRawEvents").Where("recordID", "==", r.RecordID).OrderBy(db.DocumentID, db.Asc))
		if e != nil {
			return nil, e
		}
		rawDocs = append(rawDocs, docs...)
	}
	eventsByRecord := map[string][]attendanceevent.Event{}
	for _, doc := range rawDocs {
		var raw struct {
			EventID         string    `firestore:"eventID"`
			Action          string    `firestore:"action"`
			OccurredAt      time.Time `firestore:"occurredAt"`
			RecordID        string    `firestore:"recordID"`
			ResultCode      string    `firestore:"resultCode"`
			Sequence        int64     `firestore:"sequence"`
			Source          string    `firestore:"source"`
			ContractVersion int       `firestore:"contractVersion"`
			GroupID         string    `firestore:"groupID"`
		}
		if err := doc.DataTo(&raw); err != nil {
			return nil, err
		}
		if !attendanceservice.AcceptedCode(raw.ResultCode) {
			continue
		}
		eventsByRecord[raw.RecordID] = append(eventsByRecord[raw.RecordID], attendanceevent.Event{ID: raw.EventID, Kind: raw.Action, At: raw.OccurredAt, Sequence: raw.Sequence, ResultCode: raw.ResultCode, Source: raw.Source, ContractVersion: raw.ContractVersion, RecordID: raw.RecordID, GroupID: raw.GroupID})
	}
	for i := range records {
		record := &records[i]
		if record.Revision > 0 || record.CorrectionID != "" {
			valid, err := s.verifyCorrection(ctx, user, *record)
			if err != nil {
				return nil, err
			}
			if !valid {
				record.RestConfirmed = false
				record.RebuildIssues = []string{"invalid_correction_evidence"}
			}
			continue
		}
		if record.SchemaVersion < 3 {
			continue
		} // Legacy records have no complete raw-event history.
		recordGroup := group
		if recordGroup == "" {
			recordGroup = record.GroupID
		}
		validation := attendanceservice.ValidateEventStream(eventsByRecord[record.RecordID], record.RecordID, recordGroup, record.EventSequence)
		sessions, issues := attendanceservice.RebuildSessions(eventsByRecord[record.RecordID])
		issues = append(issues, validation...)
		if len(sessions) != 1 || len(issues) > 0 || int64(len(eventsByRecord[record.RecordID])) != record.EventSequence {
			record.RestConfirmed = false
			record.RebuildIssues = []string{"incomplete_event_history"}
			for _, issue := range issues {
				record.RebuildIssues = append(record.RebuildIssues, issue.Code)
			}
			continue
		}
		rebuilt := sessions[0]
		if !record.ClockInAt.Equal(rebuilt.Start) || !record.ClockOutAt.Equal(rebuilt.End) {
			record.RestConfirmed = false
			record.RebuildIssues = []string{"projection_time_mismatch"}
			continue
		}
		record.ClockInAt, record.ClockOutAt = rebuilt.Start, rebuilt.End
		record.Breaks, record.RestConfirmed = rebuilt.Breaks, rebuilt.RestConfirmed
		record.SourceEventIDs = rebuilt.EventIDs
		record.BreakStartedAt = time.Time{}
	}
	return records, nil
}

func (s *Attendance) Recent(ctx context.Context, user string) (_ []attendanceentity.Record, err error) {
	defer func() { err = translateError(err, nil) }()
	docs, err := s.user(user).Collection("attendance").OrderBy("clockInAt", db.Desc).Limit(5).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	records := make([]attendanceentity.Record, 0, len(docs))
	for _, doc := range docs {
		var record attendanceentity.Record
		if err := decodeRecord(doc, &record); err != nil {
			return nil, err
		}
		if record.RecordID == "" {
			record.RecordID = doc.Ref.ID
		}
		records = append(records, record)
	}
	return records, nil
}
