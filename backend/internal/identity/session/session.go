// Package session manages sliding-lifetime, opaque browser sessions.
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"
)

var (
	ErrInvalid = errors.New("invalid session")
	ErrCSRF    = errors.New("invalid session csrf")
)

const (
	Lifetime         = 30 * time.Minute
	AbsoluteLifetime = 8 * time.Hour
	RenewalInterval  = 5 * time.Minute
	RenewalThreshold = 10 * time.Minute
)

type Session struct {
	User                             identity.User
	ChannelID                        string
	CreatedAt, ExpiresAt             time.Time
	CSRFHash                         string
	AbsoluteExpiresAt, LastRenewedAt time.Time
	Revision                         int64
}

type Store interface {
	Replace(ctx context.Context, previous, key string, value Session) error
	Get(ctx context.Context, key string) (Session, error)
	Delete(ctx context.Context, key string) error
	Renew(ctx context.Context, key, channel, csrf string, now func() time.Time) (Session, error)
}

type Service struct {
	Store     Store
	ChannelID string
	Now       func() time.Time
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func Key(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ValidToken(token string) bool {
	if len(token) != 43 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == token
}

func randomToken() string {
	var raw [32]byte
	// crypto/rand.Read always fills the buffer or terminates the program.
	_, _ = rand.Read(raw[:])
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

func (s *Service) Create(ctx context.Context, user identity.User, previous string) (Session, string, string, error) {
	now := s.now()
	absolute := now.Add(AbsoluteLifetime)
	if user.ExpiresAt.Before(absolute) {
		absolute = user.ExpiresAt
	}
	expires := minTime(now.Add(Lifetime), absolute)
	if user.ID == "" || !expires.After(now) || s.ChannelID == "" {
		return Session{}, "", "", ErrInvalid
	}
	token, csrf := randomToken(), randomToken()
	value := Session{User: user, ChannelID: s.ChannelID, CreatedAt: now, ExpiresAt: expires, CSRFHash: Key(csrf), AbsoluteExpiresAt: absolute, LastRenewedAt: now, Revision: 1}
	oldKey := ""
	if ValidToken(previous) {
		oldKey = Key(previous)
	}
	if err := s.Store.Replace(ctx, oldKey, Key(token), value); err != nil {
		return Session{}, "", "", err
	}
	return value, token, csrf, nil
}

func (s *Service) Load(ctx context.Context, token string) (Session, error) {
	if !ValidToken(token) {
		return Session{}, ErrInvalid
	}
	value, err := s.Store.Get(ctx, Key(token))
	if err != nil {
		return Session{}, err
	}
	if !value.ValidAt(s.now(), s.ChannelID) {
		return Session{}, ErrInvalid
	}
	return value, nil
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	if !ValidToken(token) {
		return ErrInvalid
	}
	return s.Store.Delete(ctx, Key(token))
}

func (s Session) CheckCSRF(token string) bool {
	return ValidToken(token) && subtle.ConstantTimeCompare([]byte(s.CSRFHash), []byte(Key(token))) == 1
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func (s Session) ValidAt(now time.Time, channel string) bool {
	return s.User.ID != "" && s.ChannelID == channel && s.ExpiresAt.After(now) &&
		(s.AbsoluteExpiresAt.IsZero() || s.AbsoluteExpiresAt.After(now))
}

// Renew is pure so transaction retries cannot trigger external side effects.
func (s Session) Renew(now time.Time, channel, csrf string) (Session, bool, error) {
	if !s.ValidAt(now, channel) {
		return Session{}, false, ErrInvalid
	}
	if !s.CheckCSRF(csrf) {
		return Session{}, false, ErrCSRF
	}
	// Version 1 documents have no original provider expiry: never extend them.
	if s.AbsoluteExpiresAt.IsZero() || s.LastRenewedAt.After(now) {
		return s, false, nil
	}
	if s.ExpiresAt.Sub(now) > RenewalThreshold || now.Sub(s.LastRenewedAt) < RenewalInterval {
		return s, false, nil
	}
	candidate := minTime(now.Add(Lifetime), s.AbsoluteExpiresAt)
	if !candidate.After(s.ExpiresAt) {
		return s, false, nil
	}
	s.ExpiresAt = candidate
	s.LastRenewedAt = now
	s.Revision++
	return s, true, nil
}

func (s *Service) Refresh(ctx context.Context, token, csrf string) (Session, error) {
	if !ValidToken(token) {
		return Session{}, ErrInvalid
	}
	if !ValidToken(csrf) {
		return Session{}, ErrCSRF
	}
	// Evaluate the clock inside every transaction attempt, rather than before a lock wait.
	return s.Store.Renew(ctx, Key(token), s.ChannelID, csrf, s.now)
}
