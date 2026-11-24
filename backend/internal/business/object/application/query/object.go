package query

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

func List(ctx context.Context, store port.Store, r repository.Repository, actor string, cursor string, parent ...string) (dto.Page, error) {
	if len(cursor) > 1024 || (len(parent) > 0 && parent[0] != "" && !repository.ValidID(parent[0])) {
		return dto.Page{}, entity.ErrInput
	}
	return store.List(ctx, r, actor, cursor, parent...)
}

func Get(ctx context.Context, store port.Store, r repository.Repository, id string) (entity.Object, error) {
	if !repository.ValidID(id) {
		return entity.Object{}, entity.ErrNotFound
	}
	return store.Get(ctx, r, id)
}

func ListQuantityChanges(ctx context.Context, history port.QuantityHistory, r repository.Repository, actor string, cursor string) (dto.QuantityChangePage, error) {
	if len(cursor) > 1024 {
		return dto.QuantityChangePage{}, entity.ErrInput
	}
	if history == nil {
		return dto.QuantityChangePage{}, entity.ErrUnavailable
	}
	return history.ListQuantityChanges(ctx, r, actor, cursor)
}
