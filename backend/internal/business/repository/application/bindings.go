package application

import (
	"context"
	"errors"
	"time"

	repositoryApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type ResolveOwner func(context.Context, string) (string, string, error)

// Bindings owns the application lifecycle that links a LINE interaction group
// to a repository. The group remains an external binding, never an owner or an
// authorization grant.
type Bindings struct {
	Store        repositoryApplicationPort.BindingStore
	ResolveOwner ResolveOwner
}

func (b *Bindings) ObservePersonal(ctx context.Context, owner, name, eventID string, at time.Time) error {
	if b == nil || b.Store == nil {
		return repositoryDomainEntity.ErrUnavailable
	}
	return b.Store.ObservePersonal(ctx, repositoryDomainEntity.PersonalObservation{OwnerID: owner, Name: name, EventID: eventID, At: at})
}

func (b *Bindings) ObserveGroup(ctx context.Context, actor, group, name, status, eventID string, at time.Time, users []string) error {
	if b == nil || b.Store == nil {
		return repositoryDomainEntity.ErrUnavailable
	}
	observation := repositoryDomainEntity.GroupObservation{GroupID: group, Name: name, Status: status, EventID: eventID, At: at, Users: users}
	if status == "ACTIVE" {
		if b.ResolveOwner == nil {
			observation.OwnerType, observation.OwnerID = repositoryDomainEntity.OwnerUser, actor
		} else {
			ownerType, ownerID, err := b.ResolveOwner(ctx, actor)
			if err != nil {
				return err
			}
			observation.OwnerType, observation.OwnerID = repositoryDomainEntity.OwnerType(ownerType), ownerID
		}
	}
	return b.Store.ObserveGroup(ctx, observation)
}

func (b *Bindings) Repository(ctx context.Context, group string) (repositoryDomainEntity.Repository, error) {
	if b == nil || b.Store == nil {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrUnavailable
	}
	return b.Store.GetByGroupID(ctx, group)
}

func (b *Bindings) NeedsName(ctx context.Context, group string) (bool, error) {
	current, err := b.Repository(ctx, group)
	if errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return current.Name == "", nil
}
