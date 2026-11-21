// Package application implements the private Bot invitation use cases.
package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
)

var (
	ErrUnavailableInvitation = errors.New("invitation unavailable")
	ErrForbidden             = errors.New("membership invitation forbidden")
	ErrAlreadyMember         = errors.New("already an organization member")
)

const invitationLifetime = 24 * time.Hour

type Store interface {
	Run(context.Context, func(Work) error) error
}

type Work interface {
	Organization(id, userID string) (ownerID string, ownerVersion int64, isMember bool, err error)
	Invitation(key string) (domain.Invitation, bool, error)
	CreateInvitation(key string, invitation domain.Invitation) error
	AcceptInvitation(key string, invitation domain.Invitation, at time.Time) error
	RevokeInvitation(key string) error
	Roster(organizationID string) (Roster, error)
	Departure(organizationID, actorID, requestID string) (Departure, bool, error)
	SaveDeparture(organizationID, targetID string, departure Departure) error
	LastDeparture(organizationID, actorID string) (time.Time, error)
}

type Service struct {
	Store Store
	// Code derives an opaque code from trusted channel, actor and event facts.
	Code func(actorID, eventID string) string
}

func (s Service) Invite(ctx context.Context, actor, event, organization string) (string, error) {
	code := s.Code(actor, event)
	key := digest(code)
	err := s.Store.Run(ctx, func(w Work) error {
		owner, ownerVersion, _, err := w.Organization(organization, actor)
		if err != nil {
			return err
		}
		if owner == "" || owner != actor {
			return ErrForbidden
		}
		prior, exists, err := w.Invitation(key)
		if err != nil {
			return err
		}
		now := time.Now()
		if exists {
			if prior.OrganizationID != organization || !prior.CanAccept(owner, ownerVersion, now) {
				return ErrUnavailableInvitation
			}
			return nil
		}
		return w.CreateInvitation(key, domain.Invitation{
			OrganizationID: organization, IssuerID: actor, OwnerVersion: ownerVersion, ExpiresAt: now.Add(invitationLifetime),
		})
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s Service) Accept(ctx context.Context, actor, event, code string) error {
	key := digest(code)
	return s.Store.Run(ctx, func(w Work) error {
		invitation, exists, err := w.Invitation(key)
		if err != nil {
			return err
		}
		if !exists {
			return ErrUnavailableInvitation
		}
		owner, ownerVersion, member, err := w.Organization(invitation.OrganizationID, actor)
		if err != nil {
			return err
		}
		if invitation.AcceptedBy == actor && invitation.AcceptedEvent == event && member {
			return nil
		}
		now := time.Now()
		if !invitation.CanAccept(owner, ownerVersion, now) {
			return ErrUnavailableInvitation
		}
		if member {
			return ErrAlreadyMember
		}
		departedAt, err := w.LastDeparture(invitation.OrganizationID, actor)
		if err != nil {
			return err
		}
		if !departedAt.IsZero() && !invitation.ExpiresAt.Add(-invitationLifetime).After(departedAt) {
			return ErrUnavailableInvitation
		}
		invitation.AcceptedBy, invitation.AcceptedEvent = actor, event
		return w.AcceptInvitation(key, invitation, now)
	})
}

func (s Service) Revoke(ctx context.Context, actor, code string) error {
	key := digest(code)
	return s.Store.Run(ctx, func(w Work) error {
		i, exists, err := w.Invitation(key)
		if err != nil {
			return err
		}
		if !exists {
			return ErrUnavailableInvitation
		}
		owner, _, _, err := w.Organization(i.OrganizationID, actor)
		if err != nil {
			return err
		}
		if owner == "" || owner != actor {
			return ErrForbidden
		}
		if i.Revoked {
			return nil
		}
		if i.AcceptedBy != "" || !time.Now().Before(i.ExpiresAt) {
			return ErrUnavailableInvitation
		}
		return w.RevokeInvitation(key)
	})
}

func digest(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
