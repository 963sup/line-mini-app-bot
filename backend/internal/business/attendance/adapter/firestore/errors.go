package firestore

import (
	"context"
	"errors"

	attendance "github.com/963sup/line-mini-app-bot/internal/business/attendance/application"
	boardmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/board"
	correctionsmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/corrections"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/leave"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/records"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func translateError(err, missing error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{boardmodel.ErrBoardLimit, leave.ErrInvalidLeave, leave.ErrLeaveConflict, correctionsmodel.ErrCorrection, correctionsmodel.ErrIdempotencyConflict, attendance.ErrInvalidRequest, attendance.ErrExpiredLaunch, recordmodel.ErrInvalidCursor, recordmodel.ErrRecordNotFound, attendance.ErrUnavailable} {
		if errors.Is(err, known) {
			return err
		}
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	switch status.Code(err) {
	case codes.Canceled:
		return context.Canceled
	case codes.DeadlineExceeded:
		return context.DeadlineExceeded
	}
	if missing != nil && status.Code(err) == codes.NotFound {
		return missing
	}
	return err
}
