package dto

import "github.com/963sup/line-mini-app-bot/internal/business/discussion/domain/entity"

type (
	Page struct {
		Items      []entity.Discussion
		NextCursor string
	}
	CommentPage struct {
		Items      []entity.Comment
		NextCursor string
	}
	Filter struct{ Cursor, Query string }
)
