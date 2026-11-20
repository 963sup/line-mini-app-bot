package dto

import (
	subobjectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/subobject/domain/entity"
)

type Page struct {
	Items      []subobjectDomainEntity.Subobject
	NextCursor string
}
