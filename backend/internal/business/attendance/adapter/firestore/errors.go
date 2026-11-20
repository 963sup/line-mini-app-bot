package firestore

import (
	"context"
	"errors"

	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"
	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func translateError(err, missing error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{attendancedto.ErrBoardLimit, attendanceentity.ErrInvalidLeave, attendanceentity.ErrLeaveConflict, attendanceentity.ErrCorrection, attendanceentity.ErrIdempotencyConflict, attendanceport.ErrInvalidRequest, attendanceport.ErrExpiredLaunch, attendancedto.ErrInvalidCursor, attendanceentity.ErrRecordNotFound, attendanceport.ErrUnavailable} {
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
