package application

import (
	"context"
	"time"

	repositoryApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	repositoryApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type Service struct {
	Store         repositoryApplicationPort.Store
	AddressWriter repositoryApplicationPort.AddressWriter
	OwnerAccess   repositoryApplicationPort.OwnerAccess
}

func (s *Service) UpdateAddress(ctx context.Context, actor, id string, expectedVersion int64, address repositoryDomainEntity.Address, now time.Time) (repositoryDomainEntity.Repository, error) {
	current, err := s.Get(ctx, actor, id)
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	updated, err := current.ChangeAddress(expectedVersion, address, now)
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	if s.AddressWriter == nil {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrUnavailable
	}
	return s.AddressWriter.UpdateAddress(ctx, current, updated)
}

func (s *Service) Get(ctx context.Context, user, id string) (repositoryDomainEntity.Repository, error) {
	if !repositoryDomainEntity.ValidID(id) || user == "" {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	r, err := s.Store.GetByID(ctx, id)
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	allowed, err := s.allowed(ctx, user, r)
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	if !allowed {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, user, cursor, ownerID string) (repositoryApplicationDto.Page, error) {
	if user == "" || (ownerID != "" && !repositoryDomainEntity.ValidOwnerID(ownerID)) {
		return repositoryApplicationDto.Page{}, repositoryDomainEntity.ErrNotFound
	}
	page, err := s.Store.Candidates(ctx, user, cursor, ownerID)
	if err != nil {
		return repositoryApplicationDto.Page{}, err
	}
	items := make([]repositoryDomainEntity.Repository, 0, len(page.Items))
	for _, r := range page.Items {
		if ownerID != "" && r.OwnerID != ownerID {
			continue
		}
		allowed, err := s.allowed(ctx, user, r)
		if err != nil {
			return repositoryApplicationDto.Page{}, err
		}
		if allowed {
			items = append(items, r)
		}
	}
	page.Items = items
	return page, nil
}

func (s *Service) allowed(ctx context.Context, user string, r repositoryDomainEntity.Repository) (bool, error) {
	if r.Status != "ACTIVE" {
		return false, nil
	}
	if !repositoryDomainEntity.ValidTypedOwner(r.OwnerType, r.OwnerID) {
		return false, nil
	}
	if r.OwnerType == repositoryDomainEntity.OwnerUser {
		return r.OwnerID == user, nil
	}
	if s.OwnerAccess == nil {
		return false, repositoryDomainEntity.ErrUnavailable
	}
	return s.OwnerAccess.CanAccessOwner(ctx, string(r.OwnerType), r.OwnerID, user)
}

// OwnerID returns the owner visible to the actor through the normal read path.
func (s *Service) OwnerID(ctx context.Context, actor, id string) (string, error) {
	item, err := s.Get(ctx, actor, id)
	if err != nil {
		return "", err
	}
	return item.OwnerID, nil
}
