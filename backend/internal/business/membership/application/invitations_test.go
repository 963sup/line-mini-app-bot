package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
)

type invitationRetryStore func(func(Work) error) error

func (s invitationRetryStore) Run(_ context.Context, fn func(Work) error) error { return s(fn) }

type invitationAttempt struct {
	Work
	expiresAt time.Time
	savedAt   time.Time
}

func (*invitationAttempt) Organization(string, string) (string, int64, bool, error) {
	return "owner", 1, false, nil
}

func (w *invitationAttempt) Invitation(string) (domain.Invitation, bool, error) {
	return domain.Invitation{OrganizationID: "org", IssuerID: "owner", OwnerVersion: 1, ExpiresAt: w.expiresAt}, true, nil
}

func (*invitationAttempt) LastDeparture(string, string) (time.Time, error) {
	return time.Time{}, nil
}

func (w *invitationAttempt) AcceptInvitation(_ string, _ domain.Invitation, at time.Time) error {
	w.savedAt = at
	return nil
}

func TestAcceptUsesReceiptTimeAcrossTransactionAttempts(t *testing.T) {
	expiresAt := time.Now().Add(-time.Hour)
	receivedAt := expiresAt.Add(-time.Second)
	attempts := []*invitationAttempt{{expiresAt: expiresAt}, {expiresAt: expiresAt}}
	service := Service{Store: invitationRetryStore(func(fn func(Work) error) error {
		for _, attempt := range attempts {
			if err := fn(attempt); err != nil {
				return err
			}
		}
		return nil
	})}

	if err := service.acceptAt(t.Context(), "member", "event", "code", receivedAt); err != nil {
		t.Fatal(err)
	}
	for i, attempt := range attempts {
		if !attempt.savedAt.Equal(receivedAt) {
			t.Errorf("attempt %d used %v, want %v", i+1, attempt.savedAt, receivedAt)
		}
	}
	if err := service.acceptAt(t.Context(), "member", "event", "code", expiresAt); !errors.Is(err, ErrUnavailableInvitation) {
		t.Fatalf("accept at expiration: got %v, want %v", err, ErrUnavailableInvitation)
	}
}
