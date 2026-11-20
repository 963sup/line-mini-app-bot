package milestones

import (
	"context"
	repositoryApplicationDtoMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/milestones"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/milestones"
)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Store interface {
	ListMilestones(context.Context, repositoryDomainEntity.Repository, string, string) (repositoryApplicationDtoMilestones.Page, error)
	GetMilestone(context.Context, repositoryDomainEntity.Repository, string) (repositoryDomainEntityMilestones.Milestone, error)
	CreateMilestone(context.Context, repositoryDomainEntity.Repository, string, string, repositoryDomainEntityMilestones.Content) (repositoryDomainEntityMilestones.Milestone, error)
	UpdateMilestone(context.Context, repositoryDomainEntity.Repository, string, string, repositoryDomainEntityMilestones.Update) (repositoryDomainEntityMilestones.Milestone, error)
}
