package wiki

import (
	"context"
	"errors"
	"time"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	wikimodel "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"

	repositoryApplicationPortWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/wiki"
)

type Service struct {
	Repositories repositoryApplicationPortWiki.Repositories
	Store        repositoryApplicationPortWiki.Store
}

func (s *Service) scope(ctx context.Context, actor, id string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrUnavailable
	}
	repo, err := s.Repositories.Get(ctx, actor, id)
	if errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	return repo, err
}

func (s *Service) Get(ctx context.Context, actor, repositoryID, pageID string) (wikimodel.Page, error) {
	repo, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return wikimodel.Page{}, err
	}
	return s.Store.GetWiki(ctx, repo, pageID)
}

func (s *Service) List(ctx context.Context, actor, repositoryID, pageID, cursor string) ([]wikimodel.Page, string, error) {
	repo, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return nil, "", err
	}
	return s.Store.ListWiki(ctx, repo, actor, pageID, cursor)
}

func (s *Service) Save(ctx context.Context, actor, repositoryID, id, key string, expected int64, title, body string) (wikimodel.Page, error) {
	repo, err := s.scope(ctx, actor, repositoryID)
	if err != nil {
		return wikimodel.Page{}, err
	}
	title, err = wikimodel.Validate(title, body)
	if err != nil {
		return wikimodel.Page{}, err
	}
	now := time.Now().UTC()
	target := repositoryApplicationPortWiki.WriteTarget{ID: id, Actor: actor, CreationKey: key}
	return s.Store.WriteWiki(ctx, repo, target, func(state repositoryApplicationPortWiki.WriteState) (wikimodel.Page, bool, error) {
		if id != "" {
			if state.Current == nil {
				return wikimodel.Page{}, false, repositoryDomainEntity.ErrNotFound
			}
			return state.Current.Revise(expected, actor, title, body, now)
		}
		if state.Current != nil {
			if err := state.FirstRevision.CheckCreationReplay(actor, title, body); err != nil {
				return wikimodel.Page{}, false, err
			}
			return *state.Current, false, nil
		}
		return wikimodel.Page{ID: state.ID, Title: title, Body: body, AuthorID: actor, Version: 1, UpdatedAt: now}, true, nil
	})
}
