package wiki

import (
	"context"
	"time"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	wikimodel "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"

	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationPortWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/wiki"
)

type Service struct {
	Repositories *repositoryApplication.Service
	Store        repositoryApplicationPortWiki.Store
}

func (s *Service) Save(ctx context.Context, repo repositoryDomainEntity.Repository, actor, id, key string, expected int64, title, body string) (wikimodel.Page, error) {
	title, err := wikimodel.Validate(title, body)
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
