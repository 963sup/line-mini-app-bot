package query

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/group/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/group/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/group/domain/valueobject"
)

type Service struct {
	Store      port.Store
	Membership port.Membership
}

func (s *Service) Get(ctx context.Context, user, id string) (entity.Group, error) {
	if !valueobject.ValidID(id) || user == "" {
		return entity.Group{}, entity.ErrNotFound
	}
	r, err := s.Store.GetByID(ctx, id)
	if err != nil {
		return entity.Group{}, err
	}
	allowed, err := s.allowed(ctx, user, r)
	if err != nil {
		return entity.Group{}, err
	}
	if !allowed {
		return entity.Group{}, entity.ErrNotFound
	}
	return r, nil
}

func (s *Service) List(ctx context.Context, user, cursor string) (dto.Page, error) {
	if user == "" {
		return dto.Page{}, entity.ErrNotFound
	}
	page, err := s.Store.Candidates(ctx, user, cursor)
	if err != nil {
		return dto.Page{}, err
	}
	items := make([]entity.Group, 0, len(page.Items))
	for _, r := range page.Items {
		allowed, err := s.allowed(ctx, user, r)
		if err != nil {
			return dto.Page{}, err
		}
		if allowed {
			items = append(items, r)
		}
	}
	page.Items = items
	return page, nil
}

func (s *Service) allowed(ctx context.Context, user string, r entity.Group) (bool, error) {
	if r.Status != "ACTIVE" {
		return false, nil
	}
	if s.Membership == nil {
		return false, port.ErrUnavailable
	}
	return s.Membership.IsMember(ctx, r.GroupID, user)
}
