package parsing

import (
	"context"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type Producer interface {
	Candidates(context.Context, objectimportDomainEntity.Source) ([]objectimportDomainEntity.Candidate, error)
}

type Repositories interface {
	Get(context.Context, string, string) (repository.Repository, error)
}
