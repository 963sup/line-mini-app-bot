package grpcstatus

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNormalizeContextErrorPreservesStorageFailure(t *testing.T) {
	failure := status.Error(codes.PermissionDenied, "storage access denied")
	if got := NormalizeContextError(failure); got != failure {
		t.Fatalf("got %v, want original error", got)
	}
	if got := NormalizeContextError(status.Error(codes.Canceled, "cancel")); got != context.Canceled {
		t.Fatalf("unexpected cancellation: %v", got)
	}
	if got := NormalizeContextError(status.Error(codes.DeadlineExceeded, "timeout")); got != context.DeadlineExceeded {
		t.Fatalf("unexpected deadline: %v", got)
	}
}

func TestNormalizeContextErrorPreservesIdentity(t *testing.T) {
	for _, err := range []error{nil, errors.New("domain failure"), fmt.Errorf("read interrupted: %w", context.Canceled), context.DeadlineExceeded} {
		if got := NormalizeContextError(err); got != err {
			t.Fatalf("got %v, want original %v", got, err)
		}
	}
}
