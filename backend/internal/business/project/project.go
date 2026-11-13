// Package project owns the minimal repository-scoped project workflow.
package project

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
	ErrInput            = errors.New("invalid project input")
	ErrNotFound         = errors.New("project not found")
	ErrForbidden        = errors.New("only creator may edit")
	ErrConflict         = errors.New("version conflict")
	ErrCreationConflict = errors.New("creation key conflict")
	ErrUnavailable      = errors.New("project unavailable")
	creationKey         = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type (
	Project struct {
		ID, RepositoryID, Title, Body, CreatorID string
		Version                                  int64
		CreatedAt, UpdatedAt                     time.Time
		IssueCount, CompletedIssueCount          int64
	}
	Content struct{ Title, Body string }
	Update  struct {
		Content
		ExpectedVersion int64
	}
	Page struct {
		Items      []Project
		NextCursor string
	}
)

func Normalize(c Content) (Content, error) {
	c.Title = strings.TrimSpace(c.Title)
	c.Body = strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	return c, nil
}

func CheckUpdate(current Project, actor string, version int64) error {
	if actor == "" || actor != current.CreatorID {
		return ErrForbidden
	}
	if version < 1 {
		return ErrInput
	}
	if current.Version != version {
		return ErrConflict
	}
	return nil
}

func ApplyUpdate(current Project, actor string, input Update) (Project, error) {
	if e := CheckUpdate(current, actor, input.ExpectedVersion); e != nil {
		return Project{}, e
	}
	c, e := Normalize(input.Content)
	if e != nil {
		return Project{}, e
	}
	if c.Title == current.Title && c.Body == current.Body {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Project{}, ErrConflict
	}
	current.Title = c.Title
	current.Body = c.Body
	current.Version++
	return current, nil
}

type (
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		List(context.Context, repository.Repository, string, string) (Page, error)
		Get(context.Context, repository.Repository, string) (Project, error)
		Create(context.Context, repository.Repository, string, string, Content) (Project, error)
		Update(context.Context, repository.Repository, string, string, Update) (Project, error)
		ListItems(context.Context, repository.Repository, string, string, string) (ItemPage, error)
		ChangeItem(context.Context, repository.Repository, string, string, string, bool, int64) (Project, error)
	}
)

type Service struct {
	Repositories Repositories
	Store        Store
}

func (s *Service) scope(ctx context.Context, actor, id string) (repository.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repository.Repository{}, ErrUnavailable
	}
	r, e := s.Repositories.Get(ctx, actor, id)
	if errors.Is(e, repository.ErrNotFound) {
		return repository.Repository{}, ErrNotFound
	}
	return r, e
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

func (s *Service) Get(ctx context.Context, actor, rid, id string) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	if !repository.ValidID(id) {
		return Project{}, ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c Content) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	c, e = Normalize(c)
	if e != nil || !creationKey.MatchString(key) {
		return Project{}, ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, c)
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, input Update) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	if !repository.ValidID(id) {
		return Project{}, ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, id, input)
}

type (
	Item     struct{ ID, IssueID, Title, State string }
	ItemPage struct {
		Items      []Item
		NextCursor string
	}
)

func (s *Service) ListItems(ctx context.Context, actor, rid, id, cursor string) (ItemPage, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return ItemPage{}, e
	}
	if !repository.ValidID(id) || len(cursor) > 1024 {
		return ItemPage{}, ErrInput
	}
	return s.Store.ListItems(ctx, r, actor, id, cursor)
}

func (s *Service) ChangeItem(ctx context.Context, actor, rid, id, target string, add bool, version int64) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	if !repository.ValidID(id) || !repository.ValidID(target) || version < 1 {
		return Project{}, ErrInput
	}
	return s.Store.ChangeItem(ctx, r, actor, id, target, add, version)
}
