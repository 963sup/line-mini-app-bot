package labels

import (
	"context"

	repositoryApplicationDtoLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/labels"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
)

type Repositories interface {
	Get(context.Context, string, string) (repositoryDomainEntity.Repository, error)
}

type Store interface {
	ListLabels(context.Context, repositoryDomainEntity.Repository, string, string) (repositoryApplicationDtoLabels.Page, error)
	RenameLabel(context.Context, repositoryDomainEntity.Repository, string, string, string, string) (repositoryDomainEntityLabels.Label, error)
	CreateLabel(context.Context, repositoryDomainEntity.Repository, string, string, string) (repositoryDomainEntityLabels.Label, error)
}
