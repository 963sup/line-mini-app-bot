// Package subobject owns one independent child layer beneath an Object.
package subobject

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/object"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrInput            = errors.New("invalid subobject")
	ErrNotFound         = errors.New("subobject or parent not found")
	ErrForbidden        = errors.New("parent owner required")
	ErrConflict         = errors.New("subobject version conflict")
	ErrCreationConflict = errors.New("subobject intent conflict")
	ErrProgressRemoval  = errors.New("completed subobject cannot become non-work")
	ErrUnavailable      = errors.New("subobject unavailable")
	keyPattern          = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type (
	Content   struct{ Name, Type, Quantity, Unit, CompletedQuantity string }
	Subobject struct {
		Content
		ID, RepositoryID, ParentObjectID, CreatorID, Status string
		Version                                             int64
		CreatedAt, UpdatedAt                                time.Time
	}
	Update struct {
		Content
		CompletedQuantity *string
		ExpectedVersion   int64
	}
	Page struct {
		Items      []Subobject
		NextCursor string
	}
)

// Normalize reuses exact decimal and text rules without inheriting Object source fields.
func Normalize(c Content) (Content, error) {
	c.Type = strings.TrimSpace(c.Type)
	if c.Type != object.TypeWork && c.Type != object.TypeNonWork {
		return Content{}, ErrInput
	}
	n, e := object.Normalize(object.Content{Name: c.Name, Type: c.Type, Quantity: c.Quantity, Unit: c.Unit, CompletedQuantity: c.CompletedQuantity})
	if e != nil {
		return Content{}, ErrInput
	}
	return Content{Name: n.Name, Type: n.Type, Quantity: n.Quantity, Unit: n.Unit, CompletedQuantity: n.CompletedQuantity}, nil
}

func Status(c Content) (string, error) {
	if c.Type == object.TypeNonWork && c.CompletedQuantity == "" {
		return object.StatusNotStarted, nil
	}
	status, e := object.WorkStatus(c.Quantity, c.CompletedQuantity)
	if e != nil {
		return "", ErrInput
	}
	return status, nil
}

func ApplyUpdate(current Subobject, actor string, input Update) (Subobject, error) {
	if actor == "" || actor != current.CreatorID {
		return Subobject{}, ErrForbidden
	}
	c := input.Content
	c.Type = strings.TrimSpace(c.Type)
	if input.CompletedQuantity != nil {
		c.CompletedQuantity = *input.CompletedQuantity
	} else if c.Type == object.TypeWork && current.Type == object.TypeWork {
		c.CompletedQuantity = current.CompletedQuantity
	} else {
		c.CompletedQuantity = ""
	}
	if current.Type == object.TypeWork && c.Type == object.TypeNonWork && current.CompletedQuantity != "0" {
		return Subobject{}, ErrProgressRemoval
	}
	if c.Type == object.TypeWork && c.CompletedQuantity == "" {
		return Subobject{}, ErrInput
	}
	normalized, e := Normalize(c)
	if e != nil {
		return Subobject{}, e
	}
	status, e := Status(normalized)
	if e != nil || input.ExpectedVersion < 1 {
		return Subobject{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Subobject{}, ErrConflict
	}
	if current.Content == normalized && current.Status == status {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Subobject{}, ErrConflict
	}
	current.Content = normalized
	current.Status = status
	current.Version++
	return current, nil
}

type (
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		List(context.Context, repository.Repository, string, string, string) (Page, error)
		Get(context.Context, repository.Repository, string, string, string) (Subobject, error)
		Create(context.Context, repository.Repository, string, string, string, Content) (Subobject, error)
		Update(context.Context, repository.Repository, string, string, string, Update) (Subobject, error)
	}
	Service struct {
		Repositories Repositories
		Store        Store
	}
)

func (s *Service) scope(ctx context.Context, actor, rid, pid string) (repository.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repository.Repository{}, ErrUnavailable
	}
	if actor == "" || !repository.ValidID(rid) || !repository.ValidID(pid) {
		return repository.Repository{}, ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repository.ErrNotFound) {
		return repository.Repository{}, ErrNotFound
	}
	if e != nil {
		return repository.Repository{}, ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repository.ValidGroup(r.GroupID) {
		return repository.Repository{}, ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, actor, rid, pid, cursor string) (Page, error) {
	r, e := s.scope(ctx, actor, rid, pid)
	if e != nil {
		return Page{}, e
	}
	if len(cursor) > 1024 {
		return Page{}, ErrInput
	}
	return s.Store.List(ctx, r, actor, pid, cursor)
}

func (s *Service) Get(ctx context.Context, actor, rid, pid, id string) (Subobject, error) {
	r, e := s.scope(ctx, actor, rid, pid)
	if e != nil {
		return Subobject{}, e
	}
	if !repository.ValidID(id) {
		return Subobject{}, ErrNotFound
	}
	return s.Store.Get(ctx, r, actor, pid, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, pid, key string, input Content) (Subobject, error) {
	r, e := s.scope(ctx, actor, rid, pid)
	if e != nil {
		return Subobject{}, e
	}
	if !keyPattern.MatchString(key) {
		return Subobject{}, ErrInput
	}
	if strings.TrimSpace(input.Type) == object.TypeWork && input.CompletedQuantity == "" {
		input.CompletedQuantity = "0"
	}
	c, e := Normalize(input)
	if e != nil {
		return Subobject{}, e
	}
	return s.Store.Create(ctx, r, actor, pid, key, c)
}

func (s *Service) Update(ctx context.Context, actor, rid, pid, id string, input Update) (Subobject, error) {
	r, e := s.scope(ctx, actor, rid, pid)
	if e != nil {
		return Subobject{}, e
	}
	if !repository.ValidID(id) {
		return Subobject{}, ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, pid, id, input)
}
