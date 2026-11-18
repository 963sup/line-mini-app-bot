package firestore

import (
	"context"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/punch"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
)

func (s *Attendance) Punch(ctx context.Context, group, user, event, action string, now time.Time, eventTime int64) (punch.Result, error) {
	result, err := s.executePunch(ctx, punchCommand{GroupID: group, UserID: user, EventID: event, Action: action, ReceivedAt: now, EventTimeMillis: eventTime, Mode: recordPunch})
	return result, translateError(err, nil)
}

func (s *Attendance) SubmitPunch(ctx context.Context, id, user string, position recordmodel.Position, now time.Time) (punch.Result, error) {
	result, err := s.executePunchRequest(ctx, punchCommand{RequestID: id, UserID: user, Position: &position, ReceivedAt: now, Mode: recordPunch})
	return result, translateError(err, nil)
}

func (s *Attendance) ResolvePunchAttempt(ctx context.Context, id, user string, now time.Time) (punch.Result, error) {
	result, err := s.executePunchRequest(ctx, punchCommand{RequestID: id, UserID: user, ReceivedAt: now, Mode: resolvePunch})
	return result, translateError(err, nil)
}
