package dto

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type Page struct {
	Items      []repositoryDomainEntity.Repository
	NextCursor string
}
