// Package group owns the signed LINE engineering group and its membership checks.
package group

import (
	"context"
	"errors"
	"regexp"
	"time"
)

var (
	ErrNotFound    = errors.New("group not found")
	ErrUnavailable = errors.New("group unavailable")
	ErrCursor      = errors.New("invalid group cursor")
	idPattern      = regexp.MustCompile(`^C[0-9a-f]{32}$`)
)

func ValidID(id string) bool { return idPattern.MatchString(id) }

// Group represents an engineering group within one Messaging API channel.
type Group struct {
	ID, GroupID, Name, Status string
	Version                   int64
	CreatedAt, UpdatedAt      time.Time
}

// Observation is accepted only from the signature-verified webhook adapter.
// Users are discovery candidates, never an authorization grant.
type Observation struct {
	GroupID, Name, Status, EventID string
	At                             time.Time
	Users                          []string
}

type Page struct {
	Items      []Group
	NextCursor string
}

type Store interface {
	Observe(context.Context, Observation) error
	GetByID(context.Context, string) (Group, error)
	Candidates(context.Context, string, string) (Page, error)
}

type Membership interface {
	IsMember(context.Context, string, string) (bool, error)
}

type Service struct {
	Store      Store
	Membership Membership
}

func (s *Service) Get(ctx context.Context, user, id string) (Group, error) {
	if !ValidID(id) || user == "" {
		return Group{}, ErrNotFound
	}
	r, err := s.Store.GetByID(ctx, id)
	if err != nil {
		return Group{}, err
	}
	allowed, err := s.allowed(ctx, user, r)
	if err != nil {
		return Group{}, err
	}
	if !allowed {
		return Group{}, ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, user, cursor string) (Page, error) {
	if user == "" {
		return Page{}, ErrNotFound
	}
	page, err := s.Store.Candidates(ctx, user, cursor)
	if err != nil {
		return Page{}, err
	}
	items := make([]Group, 0, len(page.Items))
	for _, r := range page.Items {
		allowed, err := s.allowed(ctx, user, r)
		if err != nil {
			return Page{}, err
		}
		if allowed {
			items = append(items, r)
		}
	}
	page.Items = items
	return page, nil
}

func (s *Service) allowed(ctx context.Context, user string, r Group) (bool, error) {
	if r.Status != "ACTIVE" {
		return false, nil
	}
	if s.Membership == nil {
		return false, ErrUnavailable
	}
	return s.Membership.IsMember(ctx, r.GroupID, user)
}
