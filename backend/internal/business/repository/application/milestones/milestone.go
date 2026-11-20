package milestones

import (
	"context"
	"errors"
	"regexp"

	repositoryApplicationDtoMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/milestones"
	repositoryApplicationPortMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/milestones"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/milestones"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct {
	Repositories repositoryApplicationPortMilestones.Repositories
	Store        repositoryApplicationPortMilestones.Store
}

func (s *Service) scope(ctx context.Context, actor, id string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntityMilestones.ErrUnavailable
	}
	r, e := s.Repositories.Get(ctx, actor, id)
	if errors.Is(e, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntityMilestones.ErrNotFound
	}
	return r, e
}

func (s *Service) List(ctx context.Context, actor, rid, cursor string) (repositoryApplicationDtoMilestones.Page, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return repositoryApplicationDtoMilestones.Page{}, e
	}
	if len(cursor) > 1024 {
		return repositoryApplicationDtoMilestones.Page{}, repositoryDomainEntityMilestones.ErrInput
	}
	return s.Store.ListMilestones(ctx, r, actor, cursor)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (repositoryDomainEntityMilestones.Milestone, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return repositoryDomainEntityMilestones.Milestone{}, e
	}
	if !repositoryDomainEntity.ValidID(id) {
		return repositoryDomainEntityMilestones.Milestone{}, repositoryDomainEntityMilestones.ErrNotFound
	}
	return s.Store.GetMilestone(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c repositoryDomainEntityMilestones.Content) (repositoryDomainEntityMilestones.Milestone, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return repositoryDomainEntityMilestones.Milestone{}, e
	}
	c, e = repositoryDomainEntityMilestones.Normalize(c)
	if e != nil || !creationKey.MatchString(key) {
		return repositoryDomainEntityMilestones.Milestone{}, repositoryDomainEntityMilestones.ErrInput
	}
	return s.Store.CreateMilestone(ctx, r, actor, key, c)
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, input repositoryDomainEntityMilestones.Update) (repositoryDomainEntityMilestones.Milestone, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return repositoryDomainEntityMilestones.Milestone{}, e
	}
	if !repositoryDomainEntity.ValidID(id) {
		return repositoryDomainEntityMilestones.Milestone{}, repositoryDomainEntityMilestones.ErrNotFound
	}
	return s.Store.UpdateMilestone(ctx, r, actor, id, input)
}
