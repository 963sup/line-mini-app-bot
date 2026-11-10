package firestore

import (
	"context"
	"fmt"
	"sort"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/attendance/worktime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Attendance struct {
	Client    *db.Client
	ChannelID string
}

func (s *Attendance) user(user string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("users").Doc(documentKey(user))
}

// Punch atomically records an event result, the shift and its active pointer.
// No external calls belong inside this retryable transaction.
func (s *Attendance) punchOperation(ctx context.Context, group, user, event, action string, now time.Time, eventTime int64) (attendance.Result, error) {
	return s.punch(ctx, group, user, event, action, now, eventTime, nil, "")
}

func (s *Attendance) punch(ctx context.Context, group, user, event, action string, now time.Time, eventTime int64, position *attendance.Position, requestID string) (attendance.Result, error) {
	u := s.user(user)
	state := u.Collection("attendanceStates").Doc(documentKey(group))
	receipt := u.Collection("attendanceEvents").Doc(documentKey(group, event, action))
	var result attendance.Result
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = attendance.Result{}
		var request requestDocument
		if requestID != "" {
			snap, err := tx.Get(s.requests().Doc(requestID))
			if err != nil {
				return err
			}
			if err := snap.DataTo(&request); err != nil {
				return err
			}
			if request.UserID != documentKey(user) || request.GroupID != group || request.EventID != event || request.Action != action {
				return attendance.ErrInvalidRequest
			}
		}
		old, err := tx.Get(receipt)
		if err == nil {
			return decodeResult(old, &result)
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		if requestID != "" && !now.Before(request.ExpiresAt) {
			return attendance.ErrInvalidRequest
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
		if eventTime < active.LastEventTime {
			result.Code = "stale_event"
			if err := tx.Create(u.Collection("attendanceRawEvents").Doc(documentKey(group, event, action)), map[string]any{"eventID": event, "action": action, "occurredAt": now, "eventTime": eventTime, "groupID": group, "source": "server", "resultCode": result.Code}); err != nil {
				return err
			}
			return tx.Create(receipt, resultDocumentFrom(result))
		}
		var record attendance.Record
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
			if record.GroupID != group || !record.ClockOutAt.IsZero() {
				return fmt.Errorf("invalid attendance state")
			}
		}
		machine := worktime.State{}
		if active.RecordID != "" {
			machine.Active = &worktime.Session{ID: active.RecordID, Start: record.ClockInAt, End: record.ClockOutAt, Breaks: record.Breaks, RestConfirmed: record.RestConfirmed, EventIDs: record.SourceEventIDs}
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
		next, outcome := worktime.ApplyEvent(machine, worktime.Event{ID: event, Kind: action, At: now})
		result.Code = outcome.Code
		if outcome.Accepted {
			if action == "clock-in" {
				record = attendance.Record{SchemaVersion: 3, GroupID: group, RecordID: documentKey(group, event), ClockInPosition: position}
				ref = u.Collection("attendance").Doc(record.RecordID)
			}
			record.ClockInAt = outcome.Session.Start
			record.ClockOutAt = outcome.Session.End
			record.IsOpen = record.ClockOutAt.IsZero()
			record.Breaks = outcome.Session.Breaks
			record.RestConfirmed = outcome.Session.RestConfirmed
			record.BreakStartedAt = next.RestStartedAt
			record.EventSequence++
			record.LastAppliedAt = now
			record.SourceEventIDs = outcome.Session.EventIDs
			if action == "clock-out" {
				record.ClockOutPosition = position
			}
			if action == "clock-in" {
				if err := tx.Create(ref, recordDocumentFrom(record)); err != nil {
					return err
				}
			} else {
				if err := tx.Set(ref, recordDocumentFrom(record)); err != nil {
					return err
				}
			}
			if err := tx.Set(u.Collection("attendanceGroups").Doc(documentKey(group)), map[string]any{"groupID": group}); err != nil {
				return err
			}
			active.RecordID = ""
			if next.Active != nil {
				active.RecordID = record.RecordID
			}
		}
		result.Record = record

		if err := tx.Set(state, map[string]any{"recordID": active.RecordID, "lastEventTime": eventTime}); err != nil {
			return err
		}
		// A separate immutable event retains the attempted action and its outcome.
		raw := map[string]any{"eventID": event, "action": action, "occurredAt": now, "source": "server", "eventTime": eventTime, "groupID": group, "recordID": active.RecordID, "resultCode": result.Code, "sequence": record.EventSequence, "contractVersion": 1}
		if position != nil {
			raw["position"] = positionDocumentFrom(position)
			raw["source"] = "located-web"
		}
		if result.Record.RecordID != "" {
			raw["recordID"] = result.Record.RecordID
		}
		if err := tx.Create(u.Collection("attendanceRawEvents").Doc(documentKey(group, event, action)), raw); err != nil {
			return err
		}
		return tx.Create(receipt, resultDocumentFrom(result))
	})
	return result, err
}

func (s *Attendance) requests() *db.CollectionRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("punchRequests")
}

func (s *Attendance) submitOperation(ctx context.Context, id, user string, position attendance.Position, now time.Time) (attendance.Result, error) {
	snapshot, err := s.requests().Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return attendance.Result{}, attendance.ErrInvalidRequest
	}
	if err != nil {
		return attendance.Result{}, err
	}
	var request requestDocument
	if err := snapshot.DataTo(&request); err != nil {
		return attendance.Result{}, err
	}
	if request.UserID != documentKey(user) {
		return attendance.Result{}, attendance.ErrInvalidRequest
	}
	return s.punch(ctx, request.GroupID, user, request.EventID, request.Action, now, request.EventTime, &position, id)
}

// Groups retains legacy reads until the offline query migration completes.
func (s *Attendance) groupsOperation(ctx context.Context, user string) ([]string, error) {
	mode, err := s.QueryMode(ctx)
	if err != nil {
		return nil, err
	}
	if mode == "legacy" {
		return s.legacyGroups(ctx, user)
	}
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

// legacyGroups is the explicitly temporary compatibility path, with historical read cost.
func (s *Attendance) legacyGroups(ctx context.Context, user string) ([]string, error) {
	docs, err := s.user(user).Collection("attendance").Select("groupID").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, doc := range docs {
		if group, ok := doc.Data()["groupID"].(string); ok && group != "" {
			seen[group] = true
		}
	}
	groups := make([]string, 0, len(seen))
	for group := range seen {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups, nil
}

func (s *Attendance) recentGroupOperation(ctx context.Context, user, group string) ([]attendance.Record, error) {
	mode, err := s.QueryMode(ctx)
	if err != nil {
		return nil, err
	}
	if mode == "indexed" {
		docs, err := s.user(user).Collection("attendance").Where("groupID", "==", group).OrderBy("clockInAt", db.Desc).OrderBy(db.DocumentID, db.Desc).Limit(5).Documents(ctx).GetAll()
		if err != nil {
			return nil, err
		}
		return decodeAttendance(docs)
	}
	records, err := s.legacyGroupRecords(ctx, user, group)
	if len(records) > 5 {
		records = records[:5]
	}
	return records, err
}

// MonthGroup includes every shift overlapping the reporting period, including open shifts.
func (s *Attendance) monthGroupOperation(ctx context.Context, user, group string, start time.Time) ([]attendance.Record, error) {
	mode, err := s.QueryMode(ctx)
	if err != nil {
		return nil, err
	}
	var records []attendance.Record
	if mode == "legacy" {
		records, err = s.legacyGroupRecords(ctx, user, group)
		if err != nil {
			return nil, err
		}
		relevant := []attendance.Record{}
		for _, r := range records {
			if overlapsMonth(r, start) {
				relevant = append(relevant, r)
			}
		}
		records = relevant
	} else {
		records, err = s.monthCandidates(ctx, user, group, start)
		if err != nil {
			return nil, err
		}
	}
	records, err = s.hydrateRecords(ctx, user, group, records)
	if err != nil {
		return nil, err
	}
	result := make([]attendance.Record, 0, len(records))
	for _, r := range records {
		if overlapsMonth(r, start) {
			result = append(result, r)
		}
	}
	return result, nil
}

// hydrateRecords preserves raw-event and correction evidence checks for every read shape.
func (s *Attendance) hydrateRecords(ctx context.Context, user, group string, records []attendance.Record) ([]attendance.Record, error) {
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
	eventsByRecord := map[string][]worktime.Event{}
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
		if !worktime.AcceptedCode(raw.ResultCode) {
			continue
		}
		eventsByRecord[raw.RecordID] = append(eventsByRecord[raw.RecordID], worktime.Event{ID: raw.EventID, Kind: raw.Action, At: raw.OccurredAt, Sequence: raw.Sequence, ResultCode: raw.ResultCode, Source: raw.Source, ContractVersion: raw.ContractVersion, RecordID: raw.RecordID, GroupID: raw.GroupID})
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
		validation := worktime.ValidateEventStream(eventsByRecord[record.RecordID], record.RecordID, group, record.EventSequence)
		sessions, issues := worktime.BuildIntervals(eventsByRecord[record.RecordID])
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

func (s *Attendance) legacyGroupRecords(ctx context.Context, user, group string) ([]attendance.Record, error) {
	// Equality-only query preserves existing data without a new composite index.
	docs, err := readQuery(ctx, s.user(user).Collection("attendance").Where("groupID", "==", group).OrderBy(db.DocumentID, db.Asc))
	if err != nil {
		return nil, err
	}
	records := make([]attendance.Record, 0, len(docs))
	for _, doc := range docs {
		var record attendance.Record
		if err := decodeRecord(doc, &record); err != nil {
			return nil, err
		}
		if record.RecordID == "" {
			record.RecordID = doc.Ref.ID
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ClockInAt.After(records[j].ClockInAt) })
	return records, nil
}

func (s *Attendance) recentOperation(ctx context.Context, user string) ([]attendance.Record, error) {
	docs, err := s.user(user).Collection("attendance").OrderBy("clockInAt", db.Desc).Limit(5).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	records := make([]attendance.Record, 0, len(docs))
	for _, doc := range docs {
		var record attendance.Record
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
