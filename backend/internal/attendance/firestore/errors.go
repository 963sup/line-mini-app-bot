package firestore

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Keep technical causes private to the adapter. Unwrap exposes only the domain category.
type storageFailure struct{ cause error }

func (e *storageFailure) Error() string { return "attendance storage unavailable" }
func (e *storageFailure) Unwrap() error { return attendance.ErrUnavailable }
func translateError(err, missing error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{attendance.ErrInvalidLeave, attendance.ErrLeaveConflict, attendance.ErrCorrection, attendance.ErrIdempotencyConflict, attendance.ErrInvalidRequest, attendance.ErrExpiredLaunch, attendance.ErrInvalidCursor, attendance.ErrRecordNotFound, attendance.ErrUnavailable} {
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
	return &storageFailure{cause: err}
}
