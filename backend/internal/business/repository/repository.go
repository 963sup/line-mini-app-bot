// Package repository owns the Bot-managed, one-group work container.
package repository

import (
	"context"
	"errors"
	"regexp"
	"time"
)

var (
	ErrNotFound    = errors.New("repository not found")
	ErrUnavailable = errors.New("repository unavailable")
	ErrCursor      = errors.New("invalid repository cursor")
	groupPattern   = regexp.MustCompile(`^C[0-9a-f]{32}$`)
	idPattern      = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

func ValidGroup(id string) bool { return groupPattern.MatchString(id) }
func ValidID(id string) bool    { return idPattern.MatchString(id) }

type Repository struct {
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
	Items      []Repository
	NextCursor string
}

type Store interface {
	Observe(context.Context, Observation) error
	GetByID(context.Context, string) (Repository, error)
	GetByGroupID(context.Context, string) (Repository, error)
	Candidates(context.Context, string, string) (Page, error)
}

type Membership interface {
	IsMember(context.Context, string, string) (bool, error)
}

type Service struct {
	Store      Store
	Membership Membership
}

func (s *Service) Get(ctx context.Context, user, id string) (Repository, error) {
	if !ValidID(id) || user == "" {
		return Repository{}, ErrNotFound
	}
	r, err := s.Store.GetByID(ctx, id)
	if err != nil {
		return Repository{}, err
	}
	allowed, err := s.allowed(ctx, user, r)
	if err != nil {
		return Repository{}, err
	}
	if !allowed {
		return Repository{}, ErrNotFound
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
	items := make([]Repository, 0, len(page.Items))
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

func (s *Service) allowed(ctx context.Context, user string, r Repository) (bool, error) {
	if r.Status != "ACTIVE" {
		return false, nil
	}
	if s.Membership == nil {
		return false, ErrUnavailable
	}
	return s.Membership.IsMember(ctx, r.GroupID, user)
}
