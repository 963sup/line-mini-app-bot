package dto

import (
	"github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
)

type (
	Target     struct{ Kind, ID, Name string }
	TargetPage struct {
		Items      []Target
		NextCursor string
	}
	Filter struct{ GroupID, WorkDate, Cursor string }
	Page   struct {
		Items      []entity.Entry
		NextCursor string
	}
)
