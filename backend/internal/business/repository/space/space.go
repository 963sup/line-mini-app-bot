// Package space owns business spaces within a repository, not LINE groups.
package space

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrInput            = errors.New("invalid repository space input")
	ErrNotFound         = errors.New("repository space not found")
	ErrCreationConflict = errors.New("repository space creation key conflict")
	ErrUnavailable      = errors.New("repository space unavailable")
	creationKey         = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type Space struct {
	ID, RepositoryID, Name, CreatorID string
	Version                           int64
	CreatedAt, UpdatedAt              time.Time
}
type Page struct {
	Items      []Space
	NextCursor string
}

func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 120 {
		return "", ErrInput
	}
	return name, nil
}

type Repositories interface {
	Get(context.Context, string, string) (repository.Repository, error)
}
type Store interface {
	List(context.Context, repository.Repository, string, string) (Page, error)
	Get(context.Context, repository.Repository, string) (Space, error)
	Create(context.Context, repository.Repository, string, string, string) (Space, error)
}
type Service struct {
	Repositories Repositories
	Store        Store
}

func (s *Service) scope(ctx context.Context, actor, rid string) (repository.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repository.Repository{}, ErrUnavailable
	}
	if actor == "" || !repository.ValidID(rid) {
		return repository.Repository{}, ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repository.ErrNotFound) {
		return repository.Repository{}, ErrNotFound
	}
	if e != nil {
		return repository.Repository{}, e
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repository.ValidGroup(r.GroupID) {
		return repository.Repository{}, ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, actor, rid, cursor string) (Page, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Page{}, e
	}
	if len(cursor) > 1024 {
		return Page{}, ErrInput
	}
	return s.Store.List(ctx, r, actor, cursor)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (Space, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Space{}, e
	}
	if !repository.ValidID(id) {
		return Space{}, ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key, name string) (Space, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Space{}, e
	}
	name, e = NormalizeName(name)
	if e != nil || !creationKey.MatchString(key) {
		return Space{}, ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, name)
}
