package repository

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/explore/application/query"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type Source struct{ Discovery application.Discovery }

func (s Source) ListVisible(ctx context.Context, actor, cursor string) (query.Page, error) {
	page, err := s.Discovery.ListDiscoverable(ctx, actor, cursor)
	if errors.Is(err, entity.ErrCursor) {
		return query.Page{}, query.ErrInput
	}
	if err != nil {
		return query.Page{}, err
	}
	result := query.Page{NextCursor: page.NextCursor}
	for _, r := range page.Items {
		result.Items = append(result.Items, query.Item{ID: r.ID, DisplayName: r.Name})
	}
	return result, nil
}
