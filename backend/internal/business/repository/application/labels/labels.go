package labels

import (
	"context"
	"errors"
	"regexp"

	repositoryApplicationDtoLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/labels"
	repositoryApplicationPortLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/labels"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct {
	Repositories repositoryApplicationPortLabels.Repositories
	Store        repositoryApplicationPortLabels.Store
}

func (s *Service) scope(ctx context.Context, actor, id string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntityLabels.ErrUnavailable
	}
	r, err := s.Repositories.Get(ctx, actor, id)
	if errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntityLabels.ErrNotFound
	}
	return r, err
}

func (s *Service) List(ctx context.Context, actor, id, cursor string) (repositoryApplicationDtoLabels.Page, error) {
	r, err := s.scope(ctx, actor, id)
	if err != nil {
		return repositoryApplicationDtoLabels.Page{}, err
	}
	if len(cursor) > 1024 {
		return repositoryApplicationDtoLabels.Page{}, repositoryDomainEntityLabels.ErrInput
	}
	return s.Store.ListLabels(ctx, r, actor, cursor)
}

func (s *Service) Create(ctx context.Context, actor, id, key, name string) (repositoryDomainEntityLabels.Label, error) {
	r, err := s.scope(ctx, actor, id)
	if err != nil {
		return repositoryDomainEntityLabels.Label{}, err
	}
	name, err = repositoryDomainEntityLabels.NormalizeName(name)
	if err != nil || !creationKey.MatchString(key) {
		return repositoryDomainEntityLabels.Label{}, repositoryDomainEntityLabels.ErrInput
	}
	return s.Store.CreateLabel(ctx, r, actor, key, name)
}

func (s *Service) Rename(ctx context.Context, actor, rid, id, expected, name string) (repositoryDomainEntityLabels.Label, error) {
	r, err := s.scope(ctx, actor, rid)
	if err != nil {
		return repositoryDomainEntityLabels.Label{}, err
	}
	name, err = repositoryDomainEntityLabels.NormalizeName(name)
	if err != nil || !repositoryDomainEntity.ValidID(id) {
		return repositoryDomainEntityLabels.Label{}, repositoryDomainEntityLabels.ErrInput
	}
	if original, e := repositoryDomainEntityLabels.NormalizeName(expected); e != nil || original != expected {
		return repositoryDomainEntityLabels.Label{}, repositoryDomainEntityLabels.ErrInput
	}
	return s.Store.RenameLabel(ctx, r, id, expected, name)
}
