// Package labels owns the minimal, immutable repository label catalog.
package labels

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrInput            = errors.New("invalid label input")
	ErrNotFound         = errors.New("repository not found")
	ErrNameConflict     = errors.New("label name already exists")
	ErrCreationConflict = errors.New("label creation key conflict")
	ErrUnavailable      = errors.New("label storage unavailable")
	creationKey         = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type (
	Label struct {
		ID, RepositoryID, Name string
		CreatedAt              time.Time
	}
	Page struct {
		Items      []Label
		NextCursor string
	}
)

func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 40 {
		return "", ErrInput
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", ErrInput
		}
	}
	return name, nil
}
func NameKey(name string) string { return strings.ToLower(name) }

// NormalizeIDs treats references as a bounded set and returns an independent slice.
func NormalizeIDs(ids []string) ([]string, error) {
	if len(ids) > 10 {
		return nil, ErrInput
	}
	result := append([]string{}, ids...)
	sort.Strings(result)
	for i, id := range result {
		if !repository.ValidID(id) || (i > 0 && id == result[i-1]) {
			return nil, ErrInput
		}
	}
	return result, nil
}

type (
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		ListLabels(context.Context, repository.Repository, string, string) (Page, error)
		CreateLabel(context.Context, repository.Repository, string, string, string) (Label, error)
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
	r, err := s.Repositories.Get(ctx, actor, id)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.Repository{}, ErrNotFound
	}
	return r, err
}

func (s *Service) List(ctx context.Context, actor, id, cursor string) (Page, error) {
	r, err := s.scope(ctx, actor, id)
	if err != nil {
		return Page{}, err
	}
	if len(cursor) > 1024 {
		return Page{}, ErrInput
	}
	return s.Store.ListLabels(ctx, r, actor, cursor)
}

func (s *Service) Create(ctx context.Context, actor, id, key, name string) (Label, error) {
	r, err := s.scope(ctx, actor, id)
	if err != nil {
		return Label{}, err
	}
	name, err = NormalizeName(name)
	if err != nil || !creationKey.MatchString(key) {
		return Label{}, ErrInput
	}
	return s.Store.CreateLabel(ctx, r, actor, key, name)
}
