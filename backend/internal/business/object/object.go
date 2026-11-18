// Package object owns repository engineering line items and their author-managed tracking.
package object

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrProgressRemoval  = errors.New("completed work cannot become non-work")
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
	SpaceID           string
	Name              string
	Description       string
	Type              string
	Quantity          string
	Unit              string
	SourceDocument    string
	SourceRevision    string
	SourceItemNo      string
	SourcePages       string
	MaterialCode      string
	CompletedQuantity string
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
		Status            string
		ExpectedVersion   int64
		CompletedQuantity *string
		SpaceID           *string
		Sources           SourceUpdate
	}
)

type Page struct {
	Items      []Object
	NextCursor string
}

const (
	TypeWork         = "work"
	TypeNonWork      = "non_work"
	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
)

func ValidStatus(s string) bool {
	switch s {
	case StatusNotStarted, StatusInProgress, StatusCompleted:
		return true
	default:
		return false
	}
}

// WorkStatus uses exact rational comparison after decimal validation.
func WorkStatus(total, done string) (string, error) {
	if !decimal.MatchString(total) || !decimal.MatchString(done) {
		return "", ErrInput
	}
	quantity, ok := new(big.Rat).SetString(total)
	if !ok || quantity.Sign() <= 0 {
		return "", ErrInput
	}
	completed, ok := new(big.Rat).SetString(done)
	if !ok || completed.Sign() < 0 || completed.Cmp(quantity) > 0 {
		return "", ErrInput
	}
	switch {
	case completed.Sign() == 0:
		return StatusNotStarted, nil
	case completed.Cmp(quantity) == 0:
		return StatusCompleted, nil
	default:
		return StatusInProgress, nil
	}
}

func canonicalDecimal(s string) string {
	if strings.Contains(s, ".") {
		return strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

func Normalize(c Content) (Content, error) {
	c.SpaceID = strings.TrimSpace(c.SpaceID)
	if c.SpaceID != "" && !repository.ValidID(c.SpaceID) {
		return Content{}, ErrInput
	}
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
	if c.CompletedQuantity != "" {
		c.CompletedQuantity = strings.TrimSpace(c.CompletedQuantity)
		if c.Type != TypeWork {
			return Content{}, ErrInput
		}
		if _, e := WorkStatus(c.Quantity, c.CompletedQuantity); e != nil {
			return Content{}, e
		}
		c.CompletedQuantity = canonicalDecimal(c.CompletedQuantity)
	}
	return c, nil
}

func ApplyUpdate(current Object, actor string, input Update) (Object, error) {
	if actor == "" || actor != current.CreatorID {
		return Object{}, ErrForbidden
	}
	content := input.Content
	content.SpaceID = current.SpaceID
	if input.SpaceID != nil {
		content.SpaceID = *input.SpaceID
	}
	content.Type = strings.TrimSpace(content.Type)
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
	// Progress is absolute, and omission preserves already managed work.
	if input.CompletedQuantity != nil {
		content.CompletedQuantity = *input.CompletedQuantity
	} else if content.Type == TypeWork && current.Type == TypeWork {
		content.CompletedQuantity = current.CompletedQuantity
	} else {
		content.CompletedQuantity = ""
	}
	if content.Type == TypeWork && current.Type != TypeWork && (input.CompletedQuantity == nil || strings.TrimSpace(*input.CompletedQuantity) == "") {
		return Object{}, ErrInput
	}
	if content.Type == TypeNonWork && input.CompletedQuantity != nil && strings.TrimSpace(*input.CompletedQuantity) != "" {
		return Object{}, ErrInput
	}
	if current.Type == TypeWork && current.CompletedQuantity != "" && content.Type != TypeWork {
		done, _ := new(big.Rat).SetString(current.CompletedQuantity)
		if done == nil || done.Sign() > 0 {
			return Object{}, ErrProgressRemoval
		}
		content.CompletedQuantity = ""
	}
	if content.Type == TypeWork && input.CompletedQuantity != nil && strings.TrimSpace(*input.CompletedQuantity) == "" {
		return Object{}, ErrInput
	}
	c, err := Normalize(content)
	status := input.Status
	switch {
	case c.Type == TypeWork && c.CompletedQuantity != "":
		derived, e := WorkStatus(c.Quantity, c.CompletedQuantity)
		if e != nil {
			return Object{}, e
		}
		if status != "" && status != derived {
			return Object{}, ErrInput
		}
		status = derived
	case status == "":
		status = current.Status
	}
	if err != nil || !ValidStatus(status) || input.ExpectedVersion < 1 {
		return Object{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Object{}, ErrConflict
	}
	if current.Content == c && current.Status == status {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Object{}, ErrConflict
	}
	current.Content = c
	current.Status = status
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
	if r.ID != id || r.Status != "ACTIVE" || !repository.ValidOwner(r.OwnerID) {
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
	if strings.TrimSpace(c.Type) == TypeWork && c.CompletedQuantity == "" {
		c.CompletedQuantity = "0"
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
