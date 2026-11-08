package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
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
func (s *Attendance) Punch(ctx context.Context, group, user, event, action string, now time.Time, eventTime int64) (attendance.Result, error) {
	return s.punch(ctx, group, user, event, action, now, eventTime, nil, "")
}

func (s *Attendance) punch(ctx context.Context, group, user, event, action string, now time.Time, eventTime int64, position *attendance.Position, requestID string) (attendance.Result, error) {
	u := s.user(user)
	state := u.Collection("attendanceStates").Doc(documentKey(group))
	receipt := u.Collection("attendanceEvents").Doc(documentKey(group, event, action))
	var result attendance.Result
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = attendance.Result{}
		var request attendance.Request
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
			return old.DataTo(&result)
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		if requestID != "" {
			if !now.Before(request.ExpiresAt) {
				return attendance.ErrInvalidRequest
			}
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
			return tx.Create(receipt, result)
		}
		var record attendance.Record
		var ref *db.DocumentRef
		if active.RecordID != "" {
			ref = u.Collection("attendance").Doc(active.RecordID)
			snapshot, err := tx.Get(ref)
			if err != nil {
				return err
			}
			if err := snapshot.DataTo(&record); err != nil {
				return err
			}
			if record.GroupID != group || !record.ClockOutAt.IsZero() {
				return fmt.Errorf("invalid attendance state")
			}
		}
		switch action {
		case "clock-in":
			if active.RecordID != "" {
				result = attendance.Result{Code: "already_in", Record: record}
				break
			}
			record = attendance.Record{SchemaVersion: 1, GroupID: group, ClockInAt: now}
			if position != nil {
				record.SchemaVersion = 2
				record.ClockInPosition = position
			}
			id := documentKey(group, event)
			if err := tx.Create(u.Collection("attendance").Doc(id), record); err != nil {
				return err
			}
			active.RecordID = id
			result = attendance.Result{Code: "clocked_in", Record: record}
		case "clock-out":
			if active.RecordID == "" {
				result.Code = "not_in"
				break
			}
			if now.Before(record.ClockInAt) {
				return fmt.Errorf("server clock precedes clock-in")
			}
			record.ClockOutAt = now
			record.ClockOutPosition = position
			if position != nil {
				record.SchemaVersion = 2
			}
			if err := tx.Set(ref, record); err != nil {
				return err
			}
			active.RecordID = ""
			result = attendance.Result{Code: "clocked_out", Record: record}
		default:
			return fmt.Errorf("invalid attendance action")
		}
		if err := tx.Set(state, map[string]any{"recordID": active.RecordID, "lastEventTime": eventTime}); err != nil {
			return err
		}
		return tx.Create(receipt, result)
	})
	return result, err
}

func (s *Attendance) requests() *db.CollectionRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("punchRequests")
}

func (s *Attendance) CreateRequest(ctx context.Context, request attendance.Request) (string, error) {
	id := documentKey(request.GroupID, request.UserID, request.EventID, request.Action)
	request.UserID = documentKey(request.UserID)
	_, err := s.requests().Doc(id).Create(ctx, request)
	if status.Code(err) == codes.AlreadyExists {
		err = nil
	}
	return id, err
}

func (s *Attendance) Submit(ctx context.Context, id, user string, position attendance.Position, now time.Time) (attendance.Result, error) {
	snapshot, err := s.requests().Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return attendance.Result{}, attendance.ErrInvalidRequest
	}
	if err != nil {
		return attendance.Result{}, err
	}
	var request attendance.Request
	if err := snapshot.DataTo(&request); err != nil {
		return attendance.Result{}, err
	}
	if request.UserID != documentKey(user) {
		return attendance.Result{}, attendance.ErrInvalidRequest
	}
	return s.punch(ctx, request.GroupID, user, request.EventID, request.Action, now, request.EventTime, &position, id)
}

// Select only group IDs, including legacy records, so no migration is needed.
func (s *Attendance) Groups(ctx context.Context, user string) ([]string, error) {
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

func (s *Attendance) RecentGroup(ctx context.Context, user, group string) ([]attendance.Record, error) {
	// Equality-only query preserves existing data without a new composite index.
	docs, err := s.user(user).Collection("attendance").Where("groupID", "==", group).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	records := make([]attendance.Record, 0, len(docs))
	for _, doc := range docs {
		var record attendance.Record
		if err := doc.DataTo(&record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ClockInAt.After(records[j].ClockInAt) })
	if len(records) > 5 {
		records = records[:5]
	}
	return records, nil
}

func documentKey(parts ...string) string {
	b, _ := json.Marshal(parts)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func (s *Attendance) Recent(ctx context.Context, user string) ([]attendance.Record, error) {
	docs, err := s.user(user).Collection("attendance").OrderBy("clockInAt", db.Desc).Limit(5).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	records := make([]attendance.Record, 0, len(docs))
	for _, doc := range docs {
		var record attendance.Record
		if err := doc.DataTo(&record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
