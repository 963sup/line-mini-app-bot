// Package issue owns repository work items and their creator-only lifecycle.
package issue

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
)

var (
	ErrInput            = errors.New("invalid issue input")
	ErrNotFound         = errors.New("issue not found")
	ErrForbidden        = errors.New("only the creator may update this issue")
	ErrConflict         = errors.New("issue version conflict")
	ErrCreationConflict = errors.New("issue creation key conflict")
	ErrUnavailable      = errors.New("issue storage unavailable")
	creationKey         = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type Issue struct {
	MilestoneID                                     string
	ID, RepositoryID, Title, Body, State, CreatorID string
	Number, Version                                 int64
	CreatedAt, UpdatedAt                            time.Time
	LabelIDs                                        []string
}

type (
	Content struct{ Title, Body string }
	Update  struct {
		Content
		State           string
		ExpectedVersion int64
		LabelIDs        *[]string
		MilestoneID     *string
	}
)

type Page struct {
	Items      []Issue
	NextCursor string
}
type Filter struct{ Cursor, State, Query string }

func ValidState(state string) bool {
	return state == "open" || state == "in_progress" || state == "completed"
}

func Normalize(c Content) (Content, error) {
	c.Title, c.Body = strings.TrimSpace(c.Title), strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	return c, nil
}

// ApplyUpdate runs inside the storage transaction against the current item.
func ApplyUpdate(current Issue, actor string, input Update) (Issue, error) {
	if actor == "" || actor != current.CreatorID {
		return Issue{}, ErrForbidden
	}
	content, err := Normalize(input.Content)
	if err != nil || !ValidState(input.State) || input.ExpectedVersion < 1 {
		return Issue{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Issue{}, ErrConflict
	}
	allowed := input.State == current.State || (current.State == "open" && input.State == "in_progress") || (current.State == "in_progress" && input.State == "completed") || (current.State == "completed" && input.State == "open")
	if !allowed {
		return Issue{}, ErrInput
	}
	ids := current.LabelIDs
	if input.LabelIDs != nil {
		ids, err = labels.NormalizeIDs(*input.LabelIDs)
		if err != nil {
			return Issue{}, ErrInput
		}
	}
	milestoneID := current.MilestoneID
	if input.MilestoneID != nil {
		milestoneID = *input.MilestoneID
		if milestoneID != "" && !repository.ValidID(milestoneID) {
			return Issue{}, ErrInput
		}
	}
	if milestoneID == current.MilestoneID && current.Title == content.Title && current.Body == content.Body && current.State == input.State && slices.Equal(current.LabelIDs, ids) {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Issue{}, ErrConflict
	}
	current.Title, current.Body, current.State = content.Title, content.Body, input.State
	current.MilestoneID = milestoneID
	current.LabelIDs = append([]string{}, ids...)
	current.Version++
	return current, nil
}

type (
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		List(context.Context, repository.Repository, string, Filter) (Page, error)
		Get(context.Context, repository.Repository, string) (Issue, error)
		Create(context.Context, repository.Repository, string, string, Content) (Issue, error)
		Update(context.Context, repository.Repository, string, string, Update) (Issue, error)
	}
)

type Service struct {
	Repositories Repositories
	Store        Store
}

func (s *Service) scope(ctx context.Context, actor, repositoryID string) (repository.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repository.Repository{}, ErrUnavailable
	}
	r, err := s.Repositories.Get(ctx, actor, repositoryID)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.Repository{}, ErrNotFound
	}
	return r, err
}

func (s *Service) List(ctx context.Context, actor, repositoryID string, filter Filter) (Page, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Page{}, err
	}
	filter.Query = strings.TrimSpace(filter.Query)
	if len(filter.Cursor) > 1024 || utf8.RuneCountInString(filter.Query) > 120 || (filter.State != "" && !ValidState(filter.State)) {
		return Page{}, ErrInput
	}
	return s.Store.List(ctx, r, actor, filter)
}

func (s *Service) Get(ctx context.Context, actor, repositoryID, id string) (Issue, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Issue{}, err
	}
	if !repository.ValidID(id) {
		return Issue{}, ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, repositoryID, key string, content Content) (Issue, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Issue{}, err
	}
	content, err = Normalize(content)
	if err != nil || !creationKey.MatchString(key) {
		return Issue{}, ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, content)
}

func (s *Service) Update(ctx context.Context, actor, repositoryID, id string, input Update) (Issue, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Issue{}, err
	}
	if !repository.ValidID(id) {
		return Issue{}, ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, id, input)
}
