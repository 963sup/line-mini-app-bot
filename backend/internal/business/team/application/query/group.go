package query

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/team/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/team/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/team/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"
)

type Service struct {
	Store      port.Store
	Membership port.Membership
	Teams      TeamQualification
}

// TeamQualification checks product membership for a group-backed team.
type TeamQualification interface {
	IsTeamMember(context.Context, string, string) (bool, error)
}

func (s *Service) Get(ctx context.Context, user, id string) (entity.Group, error) {
	if !lineid.ValidGroup(id) || user == "" {
		return entity.Group{}, entity.ErrNotFound
	}
	if s.Teams != nil {
		ok, err := s.Teams.IsTeamMember(ctx, id, user)
		if err != nil {
			return entity.Group{}, err
		}
		if !ok {
			return entity.Group{}, entity.ErrNotFound
		}
		return entity.Group{ID: id, GroupID: id}, nil
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
