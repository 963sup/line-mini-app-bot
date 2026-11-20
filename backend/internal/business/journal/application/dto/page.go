package dto

import (
	"github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
)

type (
	Filter struct{ GroupID, WorkDate, Cursor string }
	Page   struct {
		Items      []entity.Entry
		NextCursor string
	}
)
