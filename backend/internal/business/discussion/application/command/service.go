package command

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"regexp"

	"github.com/963sup/line-mini-app-bot/internal/business/discussion/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion/domain/entity"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct{ port.Access }

func (s *Service) Create(ctx context.Context, actorID, repositoryID, idempotencyKey string, content entity.Content) (entity.Discussion, error) {
	repo, err := s.Scope(ctx, actorID, repositoryID)
	if err != nil {
		return entity.Discussion{}, err
	}
	content, err = entity.Normalize(content)
	if err != nil || !creationKey.MatchString(idempotencyKey) {
		return entity.Discussion{}, entity.ErrInput
	}
	return s.Store.Create(ctx, repo, actorID, idempotencyKey, content)
}

func (s *Service) Update(ctx context.Context, actor, repositoryID, id string, input entity.Update) (entity.Discussion, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return entity.Discussion{}, err
	}
	if !repositoryDomainEntity.ValidID(id) {
		return entity.Discussion{}, entity.ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, id, input)
}

func (s *Service) CreateComment(ctx context.Context, actor, repositoryID, id, key, body string) (entity.Comment, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return entity.Comment{}, err
	}
	if !repositoryDomainEntity.ValidID(id) {
		return entity.Comment{}, entity.ErrNotFound
	}
	body, err = entity.NormalizeComment(body)
	if err != nil || !creationKey.MatchString(key) {
		return entity.Comment{}, entity.ErrInput
	}
	return s.Store.CreateComment(ctx, r, actor, id, key, body)
}

func (s *Service) UpdateComment(ctx context.Context, actor, repositoryID, id, commentID, body string, version int64) (entity.Comment, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return entity.Comment{}, err
	}
	if !repositoryDomainEntity.ValidID(id) || !repositoryDomainEntity.ValidID(commentID) {
		return entity.Comment{}, entity.ErrNotFound
	}
	return s.Store.UpdateComment(ctx, r, actor, id, commentID, body, version)
}
