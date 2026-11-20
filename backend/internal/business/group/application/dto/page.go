package dto

import (
	"github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
)

type Page struct {
	Items      []entity.Group
	NextCursor string
}
