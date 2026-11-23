package dto

import (
	"time"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

type Page struct {
	Items      []objectDomainEntity.Object
	NextCursor string
}

type QuantityChange struct {
	ObjectID, Name, Actor, BeforeQuantity, BeforeUnit, AfterQuantity, AfterUnit string
	Version                                                                     int64
	At                                                                          time.Time
}

type QuantityChangePage struct {
	Items      []QuantityChange
	NextCursor string
}
