package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
)

var (
	ErrOwnerMustTransfer = errors.New("owner must transfer ownership before leaving")
	ErrDepartureConflict = errors.New("departure request conflicts with an earlier request")
)

type Roster struct {
	Name, OwnerID string
	UserIDs       []string
}

type Member struct {
	ID                                          string
	Owner, Self, CanRemove, CanReceiveOwnership bool
}

type Members struct {
	Name                                               string
	People                                             []Member
	CanLeave, CanRemove, IsOwner, CanTransferOwnership bool
}

type Departure struct {
	ActorID, RequestID, TargetRef string
	At                            time.Time
}

func (s Service) ListMembers(ctx context.Context, actor, organization string) (Members, error) {
	var result Members
	err := s.Store.Run(ctx, func(w Work) error {
		roster, err := w.Roster(organization)
		if err != nil {
			return err
		}
		owner := roster.OwnerID == actor
		if !owner && !slices.Contains(roster.UserIDs, actor) {
			return ErrForbidden
		}
		result = Members{Name: roster.Name, CanLeave: !owner, CanRemove: owner, IsOwner: owner, CanTransferOwnership: owner}
		result.People = append(result.People, Member{ID: domain.MemberRef(organization, roster.OwnerID), Owner: true, Self: owner})
		seen := map[string]bool{roster.OwnerID: true}
		for _, user := range roster.UserIDs {
			if user == "" || seen[user] {
				continue
			}
			seen[user] = true
			result.People = append(result.People, Member{ID: domain.MemberRef(organization, user), Self: user == actor, CanRemove: owner, CanReceiveOwnership: owner})
		}
		return nil
	})
	return result, err
}

func (s Service) EndMembership(ctx context.Context, actor, organization, targetRef, requestID string) error {
	if actor == "" || requestID == "" || len(requestID) > 128 || strings.TrimSpace(requestID) != requestID {
		return ErrDepartureConflict
	}
	return s.Store.Run(ctx, func(w Work) error {
		roster, err := w.Roster(organization)
		if err != nil {
			return err
		}
		if targetRef != "" && actor != roster.OwnerID {
			return ErrForbidden
		}
		if targetRef != "" && targetRef == domain.MemberRef(organization, roster.OwnerID) {
			return ErrOwnerMustTransfer
		}
		prior, exists, err := w.Departure(organization, actor, requestID)
		if err != nil {
			return err
		}
		if exists {
			if prior.TargetRef != targetRef {
				return ErrDepartureConflict
			}
			return nil
		}
		target := actor
		if targetRef == "" {
			if actor == roster.OwnerID {
				return ErrOwnerMustTransfer
			}
			if !slices.Contains(roster.UserIDs, actor) {
				return ErrForbidden
			}
		} else {
			target = ""
			for _, user := range roster.UserIDs {
				if user != "" && domain.MemberRef(organization, user) == targetRef {
					target = user
					break
				}
			}
			if target == "" {
				return ErrForbidden
			}
		}
		return w.SaveDeparture(organization, target, Departure{actor, requestID, targetRef, time.Now()})
	})
}
