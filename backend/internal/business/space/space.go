// Package space owns the signed-group workspace identity independently of repositories.
package space

import (
	"context"
	"errors"
	"regexp"
	"time"
)

var (
	ErrNotFound    = errors.New("space not found")
	ErrUnavailable = errors.New("space unavailable")
	ErrCursor      = errors.New("invalid space cursor")
	groupPattern   = regexp.MustCompile(`^C[0-9a-f]{32}$`)
	idPattern      = regexp.MustCompile(`^C[0-9a-f]{32}$`)
)

func ValidGroup(id string) bool { return groupPattern.MatchString(id) }
func ValidID(id string) bool    { return idPattern.MatchString(id) }

type Space struct {
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
	Items      []Space
	NextCursor string
}

type Store interface {
	Observe(context.Context, Observation) error
	GetByID(context.Context, string) (Space, error)
	GetByGroupID(context.Context, string) (Space, error)
	Candidates(context.Context, string, string) (Page, error)
}

type Membership interface {
	IsMember(context.Context, string, string) (bool, error)
}

type Service struct {
	Store      Store
	Membership Membership
}

func (s *Service) Get(ctx context.Context, user, id string) (Space, error) {
	if !ValidID(id) || user == "" {
		return Space{}, ErrNotFound
	}
	r, err := s.Store.GetByID(ctx, id)
	if err != nil {
		return Space{}, err
	}
	allowed, err := s.allowed(ctx, user, r)
	if err != nil {
		return Space{}, err
	}
	if !allowed {
		return Space{}, ErrNotFound
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
	items := make([]Space, 0, len(page.Items))
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

func (s *Service) allowed(ctx context.Context, user string, r Space) (bool, error) {
	if r.Status != "ACTIVE" {
		return false, nil
	}
	if s.Membership == nil {
		return false, ErrUnavailable
	}
	return s.Membership.IsMember(ctx, r.GroupID, user)
}

// LegacyGet permits only the absent-record transition; an existing inactive or
// denied Space is never replaced by a source Repository's authorization.
func (s *Service) LegacyGet(ctx context.Context, actor, id string) (Space, error) {
	if s == nil || s.Store == nil || s.Membership == nil || !ValidID(id) || actor == "" {
		return Space{}, ErrNotFound
	}
	_, err := s.Store.GetByID(ctx, id)
	if errors.Is(err, ErrNotFound) {
		allowed, e := s.Membership.IsMember(ctx, id, actor)
		if e != nil {
			return Space{}, e
		}
		if !allowed {
			return Space{}, ErrNotFound
		}
		return Space{ID: id, GroupID: id, Status: "ACTIVE"}, nil
	}
	if err != nil {
		return Space{}, err
	}
	return s.Get(ctx, actor, id)
}
