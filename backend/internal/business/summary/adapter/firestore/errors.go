package firestore

import (
	summaryApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/summary/application/port"
)

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func translateError(err, missing error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, summaryApplicationPort.ErrUnavailable) {
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
	return err
}
