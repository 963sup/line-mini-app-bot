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

type userKey struct{}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func FromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey{}).(User)
	return user, ok && user.ID != ""
}
