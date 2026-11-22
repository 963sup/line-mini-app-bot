package dto

import (
	"github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
)

type Page struct {
	Items      []entity.Group
	NextCursor string
}
