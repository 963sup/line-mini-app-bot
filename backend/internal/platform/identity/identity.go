package identity

import (
	"context"
	"errors"
	"time"
)

type (
	User struct {
		ID, Name  string
		ExpiresAt time.Time
	}
	Verifier interface {
		Verify(context.Context, string) (User, error)
	}
)

var (
	ErrRejectedToken   = errors.New("identity token rejected")
	ErrInvalidClaims   = errors.New("invalid identity claims")
	ErrInvalidResponse = errors.New("invalid identity response")
	ErrUnavailable     = errors.New("identity service unavailable")
)

// TokenRejection carries only a safe provider status, never its response body.
type TokenRejection struct{ HTTPStatus int }

func (e *TokenRejection) Error() string { return ErrRejectedToken.Error() }
func (e *TokenRejection) Unwrap() error { return ErrRejectedToken }

type userKey struct{}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func FromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey{}).(User)
	return user, ok && user.ID != ""
}

// UserID returns the authenticated user's ID, or an empty string if absent.
func UserID(ctx context.Context) string {
	user, _ := FromContext(ctx)
	return user.ID
}
