package firestore

import (
	"context"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
)

func (s *Attendance) CreateLaunch(ctx context.Context, group, action string) (string, error) {
	value0, err := s.createLaunchOperation(ctx, group, action)
	return value0, translateError(err, nil)
}

func (s *Attendance) ExchangeLaunch(ctx context.Context, token, user string) (string, error) {
	value0, err := s.exchangeLaunchOperation(ctx, token, user)
	return value0, translateError(err, nil)
}

func (s *Attendance) LaunchGroup(ctx context.Context, token string) (string, error) {
	value0, err := s.launchGroupOperation(ctx, token)
	return value0, translateError(err, nil)
}

func (s *Attendance) Punch(ctx context.Context, group, user, event, action string, now time.Time, eventTime int64) (attendance.Result, error) {
	value0, err := s.punchOperation(ctx, group, user, event, action, now, eventTime)
	return value0, translateError(err, nil)
}

func (s *Attendance) Submit(ctx context.Context, id, user string, position attendance.Position, now time.Time) (attendance.Result, error) {
	value0, err := s.submitOperation(ctx, id, user, position, now)
	return value0, translateError(err, nil)
}

func (s *Attendance) Groups(ctx context.Context, user string) ([]string, error) {
	value0, err := s.groupsOperation(ctx, user)
	return value0, translateError(err, nil)
}

func (s *Attendance) RecentGroup(ctx context.Context, user, group string) ([]attendance.Record, error) {
	value0, err := s.recentGroupOperation(ctx, user, group)
	return value0, translateError(err, nil)
}

func (s *Attendance) MonthGroup(ctx context.Context, user, group string, start time.Time) ([]attendance.Record, error) {
	value0, err := s.monthGroupOperation(ctx, user, group, start)
	return value0, translateError(err, nil)
}

func (s *Attendance) Recent(ctx context.Context, user string) ([]attendance.Record, error) {
	value0, err := s.recentOperation(ctx, user)
	return value0, translateError(err, nil)
}

func (s *Attendance) ProposeCorrection(ctx context.Context, user string, c attendance.Correction, now time.Time) (string, error) {
	value0, err := s.proposeCorrectionOperation(ctx, user, c, now)
	return value0, translateError(err, attendance.ErrCorrection)
}

func (s *Attendance) ProposeCorrectionOnce(ctx context.Context, user, key string, c attendance.Correction, now time.Time) (string, error) {
	value0, err := s.proposeCorrectionOnceOperation(ctx, user, key, c, now)
	return value0, translateError(err, attendance.ErrCorrection)
}

func (s *Attendance) ApproveCorrection(ctx context.Context, reviewer, id string, now time.Time) error {
	return translateError(s.approveCorrectionOperation(ctx, reviewer, id, now), attendance.ErrCorrection)
}

func (s *Attendance) CorrectionStatus(ctx context.Context, user, id string) (string, error) {
	value0, err := s.correctionStatusOperation(ctx, user, id)
	return value0, translateError(err, attendance.ErrCorrection)
}

func (s *Attendance) CorrectionReview(ctx context.Context, reviewer, id string) (attendance.Correction, error) {
	value0, err := s.correctionReviewOperation(ctx, reviewer, id)
	return value0, translateError(err, attendance.ErrCorrection)
}

func (s *Attendance) GroupsPage(ctx context.Context, user, cursor string, size int) (attendance.GroupPage, error) {
	value0, err := s.groupsPageOperation(ctx, user, cursor, size)
	return value0, translateError(err, nil)
}

func (s *Attendance) RecordsPage(ctx context.Context, user, group string, start time.Time, cursor string, size int) (attendance.RecordPage, error) {
	value0, err := s.recordsPageOperation(ctx, user, group, start, cursor, size)
	return value0, translateError(err, nil)
}

func (s *Attendance) Record(ctx context.Context, user, group, id string) (attendance.Record, error) {
	value0, err := s.recordOperation(ctx, user, group, id)
	return value0, translateError(err, nil)
}
