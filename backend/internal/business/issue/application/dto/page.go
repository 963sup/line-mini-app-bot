package dto

import (
	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
)

type Page struct {
	Items      []entity.Issue
	NextCursor string
}
type Filter struct{ Cursor, State, Query string }
