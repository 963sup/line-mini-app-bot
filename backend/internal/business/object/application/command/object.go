package command

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"time"

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
	var firstErr error
	for _, item := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := store.MarkQuantityEventAttempt(ctx, item.ID); err != nil {
			return err
		}
		attempt, cancel := context.WithTimeout(ctx, 8*time.Second)
		err := errors.New("invalid stored quantity revision")
		if item.Revision.Version > 0 && item.Revision.RepositoryID != "" && item.Revision.ObjectID != "" {
			err = publish(attempt, item.ID, item.Revision)
		}
		if err == nil {
			err = store.AcknowledgeQuantityEvent(attempt, item.ID)
		}
		cancel()
		if err != nil {
			slog.WarnContext(ctx, "quantity event pending", "event_id", item.ID, "version", item.Revision.Version)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
