// Package discussion owns repository topics and independent, single-level comments.
package discussion

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
	ErrInput            = errors.New("invalid discussion input")
	ErrNotFound         = errors.New("discussion or comment not found")
	ErrForbidden        = errors.New("only the author may edit")
	ErrConflict         = errors.New("discussion version conflict")
	ErrCreationConflict = errors.New("creation key conflict")
	ErrUnavailable      = errors.New("discussion storage unavailable")
	creationKey         = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type Discussion struct {
	ID, RepositoryID, Title, Body, AuthorID string
	Number, Version                         int64
	CreatedAt, UpdatedAt                    time.Time
	LabelIDs                                []string
}
type Comment struct {
	ID, RepositoryID, DiscussionID, Body, AuthorID string
	Version                                        int64
	CreatedAt, UpdatedAt                           time.Time
}
type (
	Content struct{ Title, Body string }
	Update  struct {
		Content
		ExpectedVersion int64
		LabelIDs        *[]string
	}
	Page struct {
		Items      []Discussion
		NextCursor string
	}
	CommentPage struct {
		Items      []Comment
		NextCursor string
	}
	Filter struct{ Cursor, Query string }
)

func Normalize(c Content) (Content, error) {
	c.Title, c.Body = strings.TrimSpace(c.Title), strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	return c, nil
}

func NormalizeComment(body string) (string, error) {
	body = strings.TrimSpace(body)
	if !utf8.ValidString(body) || body == "" || utf8.RuneCountInString(body) > 4000 {
		return "", ErrInput
	}
	return body, nil
}

func ApplyUpdate(current Discussion, actor string, input Update) (Discussion, error) {
	if actor == "" || actor != current.AuthorID {
		return Discussion{}, ErrForbidden
	}
	c, err := Normalize(input.Content)
	if err != nil || input.ExpectedVersion < 1 {
		return Discussion{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Discussion{}, ErrConflict
	}
	ids := current.LabelIDs
	if input.LabelIDs != nil {
		ids, err = labels.NormalizeIDs(*input.LabelIDs)
		if err != nil {
			return Discussion{}, ErrInput
		}
	}
	if c.Title == current.Title && c.Body == current.Body && slices.Equal(current.LabelIDs, ids) {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Discussion{}, ErrConflict
	}
	current.Title, current.Body = c.Title, c.Body
	current.LabelIDs = append([]string{}, ids...)
	current.Version++
	return current, nil
}

func ApplyCommentUpdate(current Comment, actor, body string, version int64) (Comment, error) {
	if actor == "" || actor != current.AuthorID {
		return Comment{}, ErrForbidden
	}
	body, err := NormalizeComment(body)
	if err != nil || version < 1 {
		return Comment{}, ErrInput
	}
	if current.Version != version {
		return Comment{}, ErrConflict
	}
	if body == current.Body {
		return current, nil
	}
	if current.Version >= repository.MaxResourceNumber {
		return Comment{}, ErrConflict
	}
	current.Body = body
	current.Version++
	return current, nil
}

type (
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		List(context.Context, repository.Repository, string, Filter) (Page, error)
		Get(context.Context, repository.Repository, string) (Discussion, error)
		Create(context.Context, repository.Repository, string, string, Content) (Discussion, error)
		Update(context.Context, repository.Repository, string, string, Update) (Discussion, error)
		ListComments(context.Context, repository.Repository, string, string, string) (CommentPage, error)
		CreateComment(context.Context, repository.Repository, string, string, string, string) (Comment, error)
		UpdateComment(context.Context, repository.Repository, string, string, string, string, int64) (Comment, error)
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

func (s *Service) List(ctx context.Context, actor, repositoryID string, f Filter) (Page, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Page{}, err
	}
	f.Query = strings.TrimSpace(f.Query)
	if len(f.Cursor) > 1024 || !utf8.ValidString(f.Query) || utf8.RuneCountInString(f.Query) > 120 {
		return Page{}, ErrInput
	}
	return s.Store.List(ctx, r, actor, f)
}

func (s *Service) Get(ctx context.Context, actor, repositoryID, id string) (Discussion, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Discussion{}, err
	}
	if !repository.ValidID(id) {
		return Discussion{}, ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, repositoryID, key string, c Content) (Discussion, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Discussion{}, err
	}
	c, err = Normalize(c)
	if err != nil || !creationKey.MatchString(key) {
		return Discussion{}, ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, c)
}

func (s *Service) Update(ctx context.Context, actor, repositoryID, id string, input Update) (Discussion, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Discussion{}, err
	}
	if !repository.ValidID(id) {
		return Discussion{}, ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, id, input)
}

func (s *Service) ListComments(ctx context.Context, actor, repositoryID, id, cursor string) (CommentPage, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return CommentPage{}, err
	}
	if !repository.ValidID(id) {
		return CommentPage{}, ErrNotFound
	}
	if len(cursor) > 1024 {
		return CommentPage{}, ErrInput
	}
	return s.Store.ListComments(ctx, r, actor, id, cursor)
}

func (s *Service) CreateComment(ctx context.Context, actor, repositoryID, id, key, body string) (Comment, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Comment{}, err
	}
	if !repository.ValidID(id) {
		return Comment{}, ErrNotFound
	}
	body, err = NormalizeComment(body)
	if err != nil || !creationKey.MatchString(key) {
		return Comment{}, ErrInput
	}
	return s.Store.CreateComment(ctx, r, actor, id, key, body)
}

func (s *Service) UpdateComment(ctx context.Context, actor, repositoryID, id, commentID, body string, version int64) (Comment, error) {
	r, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return Comment{}, err
	}
	if !repository.ValidID(id) || !repository.ValidID(commentID) {
		return Comment{}, ErrNotFound
	}
	return s.Store.UpdateComment(ctx, r, actor, id, commentID, body, version)
}
