package query

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/discussion/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type Service struct{ port.Access }

func (s *Service) List(ctx context.Context, actor, repositoryID string, f dto.Filter) (dto.Page, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return dto.Page{}, err
	}
	f.Query = strings.TrimSpace(f.Query)
	if len(f.Cursor) > 1024 || !utf8.ValidString(f.Query) || utf8.RuneCountInString(f.Query) > 120 {
		return dto.Page{}, entity.ErrInput
	}
	return s.Store.List(ctx, r, actor, f)
}

func (s *Service) Get(ctx context.Context, actor, repositoryID, id string) (entity.Discussion, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return entity.Discussion{}, err
	}
	if !repository.ValidID(id) {
		return entity.Discussion{}, entity.ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) ListComments(ctx context.Context, actor, repositoryID, id, cursor string) (dto.CommentPage, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return dto.CommentPage{}, err
	}
	if !repository.ValidID(id) {
		return dto.CommentPage{}, entity.ErrNotFound
	}
	if len(cursor) > 1024 {
		return dto.CommentPage{}, entity.ErrInput
	}
	return s.Store.ListComments(ctx, r, actor, id, cursor)
}
