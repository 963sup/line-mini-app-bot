// Package milestones owns the minimal repository-scoped milestone workflow.
package milestones

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
	ErrInput            = errors.New("invalid milestones input")
	ErrNotFound         = errors.New("milestones not found")
	ErrForbidden        = errors.New("only creator may edit")
	ErrConflict         = errors.New("version conflict")
	ErrCreationConflict = errors.New("creation key conflict")
	ErrUnavailable      = errors.New("milestones unavailable")
	creationKey         = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type (
	Milestone struct {
		ID, RepositoryID, Title, Body, CreatorID string
		Version                                  int64
		CreatedAt, UpdatedAt                     time.Time
		IssueCount, CompletedIssueCount          int64
		State, DueDate                           string
	}
	Content struct {
		Title, Body string
		DueDate     string
	}
	Update struct {
		Content
		ExpectedVersion int64
		State           string
	}
	Page struct {
		Items      []Milestone
		NextCursor string
	}
)

func Normalize(c Content) (Content, error) {
	c.Title = strings.TrimSpace(c.Title)
	c.Body = strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	if c.DueDate != "" {
		if _, e := time.Parse("2006-01-02", c.DueDate); e != nil {
			return Content{}, ErrInput
		}
	}
	return c, nil
}

func CheckUpdate(current Milestone, actor string, version int64) error {
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

func ApplyUpdate(current Milestone, actor string, input Update) (Milestone, error) {
	if e := CheckUpdate(current, actor, input.ExpectedVersion); e != nil {
		return Milestone{}, e
	}
	c, e := Normalize(input.Content)
	if e != nil {
		return Milestone{}, e
	}
	if input.State != "open" && input.State != "completed" {
		return Milestone{}, ErrInput
	}
	if c.Title == current.Title && c.Body == current.Body && c.DueDate == current.DueDate && input.State == current.State {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Milestone{}, ErrConflict
	}
	current.Title = c.Title
	current.Body = c.Body
	current.DueDate = c.DueDate
	current.State = input.State
	current.Version++
	return current, nil
}

type (
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		ListMilestones(context.Context, repository.Repository, string, string) (Page, error)
		GetMilestone(context.Context, repository.Repository, string) (Milestone, error)
		CreateMilestone(context.Context, repository.Repository, string, string, Content) (Milestone, error)
		UpdateMilestone(context.Context, repository.Repository, string, string, Update) (Milestone, error)
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
	return s.Store.ListMilestones(ctx, r, actor, cursor)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (Milestone, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Milestone{}, e
	}
	if !repository.ValidID(id) {
		return Milestone{}, ErrNotFound
	}
	return s.Store.GetMilestone(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c Content) (Milestone, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Milestone{}, e
	}
	c, e = Normalize(c)
	if e != nil || !creationKey.MatchString(key) {
		return Milestone{}, ErrInput
	}
	return s.Store.CreateMilestone(ctx, r, actor, key, c)
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, input Update) (Milestone, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Milestone{}, e
	}
	if !repository.ValidID(id) {
		return Milestone{}, ErrNotFound
	}
	return s.Store.UpdateMilestone(ctx, r, actor, id, input)
}
