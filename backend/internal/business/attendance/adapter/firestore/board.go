package firestore

import (
	"context"
	"fmt"
	"time"

	db "cloud.google.com/go/firestore"
	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	boardUserLimit        = 200
	boardDailyRecordLimit = 500
)

// currentShift reads only the trusted user's group state and validates its target.
func currentShift(tx *db.Transaction, employee *db.DocumentRef, group string) (*attendanceentity.Record, error) {
	snap, err := tx.Get(employee.Collection("attendanceStates").Doc(firestorekey.Encode(group)))
	if status.Code(err) == codes.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var active struct {
		RecordID string `firestore:"recordID"`
	}
	if err := snap.DataTo(&active); err != nil {
		return nil, err
	}
	if active.RecordID == "" {
		return nil, nil
	}
	if !validSegment(active.RecordID) {
		return nil, fmt.Errorf("invalid active shift ID")
	}
	snap, err = tx.Get(employee.Collection("attendance").Doc(active.RecordID))
	if err != nil {
		return nil, err
	}
	var record attendanceentity.Record
	if err := decodeRecord(snap, &record); err != nil {
		return nil, err
	}
	if record.GroupID != group || !record.ClockOutAt.IsZero() || record.ClockInAt.IsZero() {
		return nil, fmt.Errorf("invalid active shift")
	}
	record.RecordID = active.RecordID
	return &record, nil
}

func (s *Attendance) CurrentShift(ctx context.Context, group, user string) (*attendanceentity.Record, error) {
	var result *attendanceentity.Record
	err := s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		var err error
		result, err = currentShift(tx, s.user(user), group)
		return err
	}, db.ReadOnly)
	return result, translateError(err, nil)
}

func (s *Attendance) GroupBoard(ctx context.Context, group string, start, now time.Time) (_ attendancedto.Board, err error) {
	defer func() { err = translateError(err, nil) }()
	// Historical user parents may not exist: DocumentRefs includes these parents.
	refs := s.Client.Collection("channels").Doc(s.ChannelID).Collection("users").DocumentRefs(ctx)
	employees := make([]*db.DocumentRef, 0)
	for {
		ref, err := refs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return attendancedto.Board{}, err
		}
		if len(employees) >= boardUserLimit {
			return attendancedto.Board{}, attendancedto.ErrBoardLimit
		}
		employees = append(employees, ref)
	}
	var result attendancedto.Board
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = attendancedto.Board{}
		for _, employee := range employees {
			active, err := currentShift(tx, employee, group)
			if err != nil {
				return err
			}
			if active != nil {
				if active.BreakStartedAt.IsZero() {
					result.Working++
				} else {
					result.Resting++
				}
				continue
			}
			// One indexed field only; filter the group in Go to avoid new composite indexes.
			q := employee.Collection("attendance").Where("clockOutAt", ">=", start).Where("clockOutAt", "<=", now).Limit(boardDailyRecordLimit + 1)
			snaps, err := tx.Documents(q).GetAll()
			if err != nil {
				return err
			}
			if len(snaps) > boardDailyRecordLimit {
				return attendancedto.ErrBoardLimit
			}
			for _, snap := range snaps {
				var r recordDocument
				if err := snap.DataTo(&r); err != nil {
					return err
				}
				if r.GroupID == group && !r.ClockInAt.IsZero() && !r.ClockOutAt.Before(r.ClockInAt) {
					result.Finished++
					break
				}
			}
		}
		return nil
	}, db.ReadOnly)
	if err != nil {
		return attendancedto.Board{}, err
	}
	result.AsOf = time.Now().UTC()
	return result, nil
}

var _ attendanceport.BoardStore = (*Attendance)(nil)
