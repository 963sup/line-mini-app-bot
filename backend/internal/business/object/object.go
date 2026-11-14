// Package object owns repository engineering line items and their author-managed tracking.
package object

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
	ErrInput            = errors.New("invalid object input")
	ErrNotFound         = errors.New("object not found")
	ErrForbidden        = errors.New("only the creator may update this object")
	ErrConflict         = errors.New("object version conflict")
	ErrCreationConflict = errors.New("object creation key conflict")
	ErrUnavailable      = errors.New("object storage unavailable")
	decimal             = regexp.MustCompile(`^(0|[1-9][0-9]{0,17})(\.[0-9]{1,6})?$`)
	creationKey         = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type Content struct {
	Name           string
	Description    string
	Type           string
	Quantity       string
	Unit           string
	SourceDocument string
	SourceRevision string
	SourceItemNo   string
	SourcePages    string
	MaterialCode   string
}
type Object struct {
	Content
	ID, RepositoryID, CreatorID, Status string
	Version                             int64
	CreatedAt, UpdatedAt                time.Time
}
type (
	SourceUpdate struct{ SourceDocument, SourceRevision, SourceItemNo, SourcePages, MaterialCode *string }
	Update       struct {
		Content
		Status          string
		ExpectedVersion int64
		Sources         SourceUpdate
	}
)

type Page struct {
	Items      []Object
	NextCursor string
}

func ValidStatus(s string) bool { return s == "not_started" || s == "in_progress" || s == "completed" }

func Normalize(c Content) (Content, error) {
	fields := []struct {
		v        *string
		limit    int
		required bool
	}{{&c.Name, 120, true}, {&c.Description, 8000, false}, {&c.Type, 80, true}, {&c.Unit, 32, true}, {&c.SourceDocument, 200, false}, {&c.SourceRevision, 80, false}, {&c.SourceItemNo, 80, false}, {&c.SourcePages, 80, false}, {&c.MaterialCode, 120, false}}
	for _, f := range fields {
		*f.v = strings.TrimSpace(*f.v)
		if !utf8.ValidString(*f.v) || utf8.RuneCountInString(*f.v) > f.limit || (f.required && *f.v == "") {
			return Content{}, ErrInput
		}
	}
	c.Quantity = strings.TrimSpace(c.Quantity)
	if !decimal.MatchString(c.Quantity) {
		return Content{}, ErrInput
	}
	if strings.Contains(c.Quantity, ".") {
		c.Quantity = strings.TrimRight(strings.TrimRight(c.Quantity, "0"), ".")
	}
	if c.Quantity == "0" {
		return Content{}, ErrInput
	}
	return c, nil
}

func ApplyUpdate(current Object, actor string, input Update) (Object, error) {
	if actor == "" || actor != current.CreatorID {
		return Object{}, ErrForbidden
	}
	content := input.Content
	for _, f := range []struct {
		value  *string
		update *string
		prior  string
	}{{&content.SourceDocument, input.Sources.SourceDocument, current.SourceDocument}, {&content.SourceRevision, input.Sources.SourceRevision, current.SourceRevision}, {&content.SourceItemNo, input.Sources.SourceItemNo, current.SourceItemNo}, {&content.SourcePages, input.Sources.SourcePages, current.SourcePages}, {&content.MaterialCode, input.Sources.MaterialCode, current.MaterialCode}} {
		if f.update == nil {
			*f.value = f.prior
		} else {
			*f.value = *f.update
		}
	}
	c, err := Normalize(content)
	if err != nil || !ValidStatus(input.Status) || input.ExpectedVersion < 1 {
		return Object{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Object{}, ErrConflict
	}
	if current.Content == c && current.Status == input.Status {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Object{}, ErrConflict
	}
	current.Content = c
	current.Status = input.Status
	current.Version++
	return current, nil
}

type Repositories interface {
	Get(context.Context, string, string) (repository.Repository, error)
}
type Store interface {
	List(context.Context, repository.Repository, string, string) (Page, error)
	Get(context.Context, repository.Repository, string) (Object, error)
	Create(context.Context, repository.Repository, string, string, Content) (Object, error)
	Update(context.Context, repository.Repository, string, string, Update) (Object, error)
}
type Service struct {
	Repositories Repositories
	Store        Store
}

func (s *Service) scope(ctx context.Context, actor, id string) (repository.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repository.Repository{}, ErrUnavailable
	}
	if actor == "" || !repository.ValidID(id) {
		return repository.Repository{}, ErrNotFound
	}
	r, err := s.Repositories.Get(ctx, actor, id)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.Repository{}, ErrNotFound
	}
	if err != nil {
		return repository.Repository{}, err
	}
	// Membership is checked by Repositories.Get; active identity/scope is still explicit here.
	if r.ID != id || r.Status != "ACTIVE" || !repository.ValidGroup(r.GroupID) {
		return repository.Repository{}, ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, actor, id, cursor string) (Page, error) {
	r, e := s.scope(ctx, actor, id)
	if e != nil {
		return Page{}, e
	}
	if len(cursor) > 1024 {
		return Page{}, ErrInput
	}
	return s.Store.List(ctx, r, actor, cursor)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Object{}, e
	}
	if !repository.ValidID(id) {
		return Object{}, ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c Content) (Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Object{}, e
	}
	c, e = Normalize(c)
	if e != nil || !creationKey.MatchString(key) {
		return Object{}, ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, c)
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, u Update) (Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Object{}, e
	}
	if !repository.ValidID(id) {
		return Object{}, ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, id, u)
}
