package application

import (
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"
	"regexp"
	"strings"

	objectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	objectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct {
	Templates          objectApplicationPort.TemplateStore
	CandidateExtractor operations.CandidateExtractor
	// DocumentTextExtractor is reserved for a future PDF import use case.
	DocumentTextExtractor operations.DocumentTextExtractor
	Repositories          objectApplicationPort.Repositories
	Store                 objectApplicationPort.Store
}

func (s *Service) scope(ctx context.Context, actor, id string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repositoryDomainEntity.Repository{}, objectDomainEntity.ErrUnavailable
	}
	if actor == "" || !repositoryDomainEntity.ValidID(id) {
		return repositoryDomainEntity.Repository{}, objectDomainEntity.ErrNotFound
	}
	r, err := s.Repositories.Get(ctx, actor, id)
	if errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, objectDomainEntity.ErrNotFound
	}
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	// Membership is checked by Repositories.Get; active identity/scope is still explicit here.
	if r.ID != id || r.Status != "ACTIVE" || !repositoryDomainEntity.ValidOwner(r.OwnerID) {
		return repositoryDomainEntity.Repository{}, objectDomainEntity.ErrNotFound
	}
	return r, nil
}

// ExtractCandidates authorizes the document request without creating objects.
func (s *Service) ExtractCandidates(ctx context.Context, actor, repositoryID string, document operations.Document) (operations.Extraction, error) {
	if _, err := s.scope(ctx, actor, repositoryID); err != nil {
		return operations.Extraction{}, err
	}
	if s.CandidateExtractor == nil {
		return operations.Extraction{}, objectDomainEntity.ErrUnavailable
	}
	return s.CandidateExtractor.ExtractCandidates(ctx, document)
}

func (s *Service) List(ctx context.Context, actor, id, cursor string, parent ...string) (objectApplicationDto.Page, error) {
	r, e := s.scope(ctx, actor, id)
	if e != nil {
		return objectApplicationDto.Page{}, e
	}
	if len(cursor) > 1024 || (len(parent) > 0 && parent[0] != "" && !repositoryDomainEntity.ValidID(parent[0])) {
		return objectApplicationDto.Page{}, objectDomainEntity.ErrInput
	}
	return s.Store.List(ctx, r, actor, cursor, parent...)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (objectDomainEntity.Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectDomainEntity.Object{}, e
	}
	if !repositoryDomainEntity.ValidID(id) {
		return objectDomainEntity.Object{}, objectDomainEntity.ErrNotFound
	}
	return s.Store.Get(ctx, r, id)
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c objectDomainEntity.Content) (objectDomainEntity.Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectDomainEntity.Object{}, e
	}
	if strings.TrimSpace(c.Type) == objectDomainEntity.TypeWork && strings.TrimSpace(c.Quantity) != "" && c.CompletedQuantity == "" {
		c.CompletedQuantity = "0"
	}
	c, e = objectDomainEntity.Normalize(c)
	if e != nil || !creationKey.MatchString(key) {
		return objectDomainEntity.Object{}, objectDomainEntity.ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, c)
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, u objectDomainEntity.Update) (objectDomainEntity.Object, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectDomainEntity.Object{}, e
	}
	if !repositoryDomainEntity.ValidID(id) {
		return objectDomainEntity.Object{}, objectDomainEntity.ErrNotFound
	}
	return s.Store.Update(ctx, r, actor, id, u)
}
