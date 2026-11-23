package command

import (
	"context"
	"regexp"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/event"
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

// DeliverQuantityEvents leaves unconfirmed publications pending for the next run.
func DeliverQuantityEvents(ctx context.Context, store port.QuantityEvents, publish func(context.Context, string, event.Revision) error) error {
	pending, err := store.PendingQuantityEvents(ctx)
	if err != nil {
		return err
	}
	for _, item := range pending {
		if err := publish(ctx, item.ID, item.Revision); err != nil {
			return err
		}
		if err := store.AcknowledgeQuantityEvent(ctx, item.ID); err != nil {
			return err
		}
	}
	return nil
}
