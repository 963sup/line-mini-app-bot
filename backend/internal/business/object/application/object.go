package application

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/command"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/query"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type Service struct {
	QuantityHistory   port.QuantityHistory
	TemplatePreparer  TemplatePreparer
	Templates         port.TemplateStore
	PersonalTemplates port.PersonalTemplateStore
	CandidatePreparer port.CandidatePreparer
	Repositories      port.Repositories
	Store             port.Store
}

func (s *Service) scope(ctx context.Context, actor, id string) (repository.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repository.Repository{}, entity.ErrUnavailable
	}
	if actor == "" || !repository.ValidID(id) {
		return repository.Repository{}, entity.ErrNotFound
	}
	r, err := s.Repositories.Get(ctx, actor, id)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.Repository{}, entity.ErrNotFound
	}
	if err != nil {
		return repository.Repository{}, err
	}
	// Membership is checked by Repositories.Get; active identity/scope is still explicit here.
	if r.ID != id || r.Status != "ACTIVE" || !repository.ValidTypedOwner(r.OwnerType, r.OwnerID) {
		return repository.Repository{}, entity.ErrNotFound
	}
	return r, nil
}

func (s *Service) ExtractCandidates(ctx context.Context, actor, repositoryID string, document dto.Document) (dto.CandidateResult, error) {
	if _, err := s.scope(ctx, actor, repositoryID); err != nil {
		return dto.CandidateResult{}, err
	}
	if s.CandidatePreparer == nil {
		return dto.CandidateResult{}, entity.ErrUnavailable
	}
	return s.CandidatePreparer.PrepareCandidates(ctx, actor, repositoryID, document)
}

func (s *Service) List(ctx context.Context, actor, id, cursor string, parent ...string) (dto.Page, error) {
	r, e := s.scope(ctx, actor, id)
	if e != nil {
		return dto.Page{}, e
	}
	return query.List(ctx, s.Store, r, actor, cursor, parent...)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (entity.Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return entity.Object{}, e
	}
	return query.Get(ctx, s.Store, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c entity.Content) (entity.Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return entity.Object{}, e
	}
	return command.Create(ctx, s.Store, r, actor, key, c)
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, u entity.Update) (entity.Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return entity.Object{}, e
	}
	return command.Update(ctx, s.Store, r, actor, id, u)
}

func (s *Service) ListQuantityChanges(ctx context.Context, actor, rid, cursor string) (dto.QuantityChangePage, error) {
	r, err := s.scope(ctx, actor, rid)
	if err != nil {
		return dto.QuantityChangePage{}, err
	}
	return query.ListQuantityChanges(ctx, s.QuantityHistory, r, actor, cursor)
}
