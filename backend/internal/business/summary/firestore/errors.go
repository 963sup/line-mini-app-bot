package firestore

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/summary"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Keep technical causes private to the adapter. Unwrap exposes only the domain category.
type storageFailure struct{ cause error }

func (e *storageFailure) Error() string { return "summary storage unavailable" }
func (e *storageFailure) Unwrap() error { return summary.ErrUnavailable }
func translateError(err, missing error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, summary.ErrUnavailable) {
		return err
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
