package dto

import (
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

type Page struct {
	Items      []objectDomainEntity.Object
	NextCursor string
}
