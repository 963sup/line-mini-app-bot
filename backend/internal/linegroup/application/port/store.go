package port

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/linegroup/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/linegroup/domain/event"
)

var (
	ErrUnavailable = errors.New("group unavailable")
	ErrCursor      = errors.New("invalid group cursor")
)

type Store interface {
	Observe(context.Context, event.Observation) error
	GetByID(context.Context, string) (entity.Group, error)
	Candidates(context.Context, string, string) (dto.Page, error)
}

type Membership interface {
	IsMember(context.Context, string, string) (bool, error)
}
