package dto

import (
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

type Page struct {
	Items      []projectDomainEntity.Project
	NextCursor string
}
