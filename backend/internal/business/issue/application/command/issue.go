package command

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"regexp"

	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Operations struct{ port.Access }

func (s *Operations) Create(ctx context.Context, actorID, repositoryID, idempotencyKey string, content entity.Content) (entity.Issue, error) {
	repo, err := s.Scope(ctx, actorID, repositoryID)
	if err != nil {
		return entity.Issue{}, err
	}
	content, err = entity.Normalize(content)
	if err != nil || !creationKey.MatchString(idempotencyKey) {
		return entity.Issue{}, entity.ErrInput
	}
	return s.Store.Create(ctx, repo, actorID, idempotencyKey, content)
}

func (s *Operations) Update(ctx context.Context, actor, repositoryID, id string, input entity.Update) (entity.Issue, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return entity.Issue{}, err
	}
	if !repositoryDomainEntity.ValidID(id) {
		return entity.Issue{}, entity.ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, id, input)
}
