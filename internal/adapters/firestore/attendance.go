package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

func key(parts ...string) string {
	b, _ := json.Marshal(parts)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func (s *Attendance) user(user string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("users").Doc(key(user))
}

// Punch atomically records an event result, the shift and its active pointer.
// No external calls belong inside this retryable transaction.
func (s *Attendance) Punch(ctx context.Context, group, user, event, action string, now time.Time, eventTime int64) (attendance.Result, error) {
	u := s.user(user)
	state := u.Collection("attendanceStates").Doc(key(group))
	receipt := u.Collection("attendanceEvents").Doc(key(group, event, action))
	var result attendance.Result
	err := s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = attendance.Result{}
		old, err := tx.Get(receipt)
		if err == nil {
			return old.DataTo(&result)
		}
		if status.Code(err) != codes.NotFound {
			return err
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
			id := key(group, event)
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
