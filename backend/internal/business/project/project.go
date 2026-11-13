// Package project owns Space projects and their source references.
package project

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/space"
)

var (
	ErrInput             = errors.New("invalid project input")
	ErrNotFound          = errors.New("project not found")
	ErrForbidden         = errors.New("only creator may edit")
	ErrConflict          = errors.New("version conflict")
	ErrCreationConflict  = errors.New("creation key conflict")
	ErrUnavailable       = errors.New("project unavailable")
	ErrSourceUnavailable = errors.New("project source unavailable")
	creationKey          = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type (
	Project struct {
		ID, SourceRepositoryID, SpaceID, Title, Body, CreatorID string
		SourceAvailable                                         bool
		Version                                                 int64
		CreatedAt, UpdatedAt                                    time.Time
		IssueCount, CompletedIssueCount                         int64
	}
	Content struct{ Title, Body, SourceRepositoryID string }
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
	if current.Version >= MaxVersion {
		return Project{}, ErrConflict
	}
	current.Title = c.Title
	current.Body = c.Body
	current.Version++
	return current, nil
}

type (
	Spaces interface {
		Get(context.Context, string, string) (space.Space, error)
		LegacyGet(context.Context, string, string) (space.Space, error)
	}
	Sources interface {
		Authorize(context.Context, string, string, string) (Source, error)
	}
	Store interface {
		List(context.Context, Scope, string, string) (Page, error)
		Get(context.Context, Scope, string) (Project, error)
		Create(context.Context, Scope, string, string, Content) (Project, error)
		Update(context.Context, Scope, string, string, Update) (Project, error)
		ListItems(context.Context, Scope, string, string, string) (ItemPage, error)
		ChangeItem(context.Context, Scope, string, string, string, bool, int64) (Project, error)
		Progress(context.Context, Scope, *Project) error
	}
)

type Service struct {
	Spaces  Spaces
	Sources Sources
	Store   Store
}

// Source is a read-only reference, never a Project authorization grant.
type (
	Source struct{ ID, SpaceID string }
	Scope  struct {
		SpaceID, SourceRepositoryID, Actor string
		Legacy                             bool
	}
)

const MaxVersion int64 = 1<<53 - 1

var idPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }
func (s *Service) scope(ctx context.Context, actor, id string) (Scope, error) {
	if s == nil || s.Spaces == nil || s.Sources == nil || s.Store == nil {
		return Scope{}, ErrUnavailable
	}
	if space.ValidID(id) {
		value, err := s.Spaces.Get(ctx, actor, id)
		if errors.Is(err, space.ErrNotFound) {
			return Scope{}, ErrNotFound
		}
		return Scope{SpaceID: value.ID, Actor: actor}, err
	}
	// A legacy route's read-only source resolver is the sole compatibility bridge.
	source, err := s.Sources.Authorize(ctx, actor, id, "")
	if err != nil {
		return Scope{}, err
	}
	value, err := s.Spaces.LegacyGet(ctx, actor, source.SpaceID)
	if errors.Is(err, space.ErrNotFound) {
		return Scope{}, ErrNotFound
	}
	if err != nil {
		return Scope{}, err
	}
	return Scope{SpaceID: value.ID, SourceRepositoryID: source.ID, Actor: actor, Legacy: true}, nil
}

func (s *Service) progress(ctx context.Context, r Scope, p *Project) error {
	_, err := s.Sources.Authorize(ctx, r.Actor, p.SourceRepositoryID, r.SpaceID)
	if err != nil {
		if r.Legacy || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		p.SourceAvailable = false
		p.IssueCount, p.CompletedIssueCount = 0, 0
		return nil
	}
	p.SourceAvailable = true
	err = s.Store.Progress(ctx, r, p)
	if !r.Legacy && errors.Is(err, ErrSourceUnavailable) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		p.SourceAvailable = false
		p.IssueCount, p.CompletedIssueCount = 0, 0
		return nil
	}
	return err
}

func (s *Service) List(ctx context.Context, actor, rid, cursor string) (Page, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Page{}, e
	}
	if len(cursor) > 1024 {
		return Page{}, ErrInput
	}
	page, err := s.Store.List(ctx, r, actor, cursor)
	if err != nil {
		return Page{}, err
	}
	for i := range page.Items {
		if err = s.progress(ctx, r, &page.Items[i]); err != nil {
			return Page{}, err
		}
	}
	return page, nil
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	if !ValidID(id) {
		return Project{}, ErrNotFound
	}
	result, err := s.Store.Get(ctx, r, id)
	if err == nil {
		err = s.progress(ctx, r, &result)
	}
	return result, err
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c Content) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	if r.Legacy {
		c.SourceRepositoryID = r.SourceRepositoryID
	}
	if !ValidID(c.SourceRepositoryID) {
		return Project{}, ErrInput
	}
	if _, e = s.Sources.Authorize(ctx, actor, c.SourceRepositoryID, r.SpaceID); e != nil {
		return Project{}, e
	}
	r.SourceRepositoryID = c.SourceRepositoryID
	c, e = Normalize(c)
	if e != nil || !creationKey.MatchString(key) {
		return Project{}, ErrInput
	}
	value, e := s.Store.Create(ctx, r, actor, key, c)
	if e == nil {
		e = s.progress(ctx, r, &value)
	}
	return value, e
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, input Update) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	if !ValidID(id) {
		return Project{}, ErrNotFound
	}
	value, e := s.Store.Update(ctx, r, actor, id, input)
	if e == nil {
		e = s.progress(ctx, r, &value)
	}
	return value, e
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
	if !ValidID(id) || len(cursor) > 1024 {
		return ItemPage{}, ErrInput
	}
	value, e := s.Store.Get(ctx, r, id)
	if e != nil {
		return ItemPage{}, e
	}
	if _, e = s.Sources.Authorize(ctx, actor, value.SourceRepositoryID, r.SpaceID); e != nil {
		return ItemPage{}, e
	}
	r.SourceRepositoryID = value.SourceRepositoryID
	return s.Store.ListItems(ctx, r, actor, id, cursor)
}

func (s *Service) ChangeItem(ctx context.Context, actor, rid, id, target string, add bool, version int64) (Project, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Project{}, e
	}
	if !ValidID(id) || !ValidID(target) || version < 1 {
		return Project{}, ErrInput
	}
	value, e := s.Store.Get(ctx, r, id)
	if e != nil {
		return Project{}, e
	}
	if _, e = s.Sources.Authorize(ctx, actor, value.SourceRepositoryID, r.SpaceID); e != nil {
		return Project{}, e
	}
	r.SourceRepositoryID = value.SourceRepositoryID
	result, e := s.Store.ChangeItem(ctx, r, actor, id, target, add, version)
	if e == nil {
		e = s.progress(ctx, r, &result)
	}
	return result, e
}
