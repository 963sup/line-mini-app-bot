package command

import (
	"context"
	"regexp"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

func ValidCreationKey(key string) bool { return creationKey.MatchString(key) }

func Create(ctx context.Context, store port.Store, r repository.Repository, actor string, key string, c entity.Content) (entity.Object, error) {
	c, e := entity.NormalizeCreation(c)
	if e != nil || !creationKey.MatchString(key) {
		return entity.Object{}, entity.ErrInput
	}
	return store.Create(ctx, r, actor, key, c)
}

func Update(ctx context.Context, store port.Store, r repository.Repository, actor string, id string, u entity.Update) (entity.Object, error) {
	if !repository.ValidID(id) {
		return entity.Object{}, entity.ErrNotFound
	}
	return store.Update(ctx, r, actor, id, u)
}
