package query

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/policy"
)

type Operations struct{ port.Access }

func (s *Operations) List(ctx context.Context, actor, repositoryID string, filter dto.Filter) (dto.Page, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return dto.Page{}, err
	}
	filter.Query = strings.TrimSpace(filter.Query)
	if len(filter.Cursor) > 1024 || utf8.RuneCountInString(filter.Query) > 120 || (filter.State != "" && !policy.ValidState(filter.State)) {
		return dto.Page{}, entity.ErrInput
	}
	return s.Store.List(ctx, r, actor, filter)
}

func (s *Operations) Get(ctx context.Context, actor, repositoryID, id string) (entity.Issue, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return entity.Issue{}, err
	}
	if !repositoryDomainEntity.ValidID(id) {
		return entity.Issue{}, entity.ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}
