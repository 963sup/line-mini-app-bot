package application

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type DiscoveryStore interface {
	DiscoverableCandidates(context.Context, string) (dto.Page, error)
}

type Discovery struct{ Store DiscoveryStore }

func (d Discovery) ListDiscoverable(ctx context.Context, actor, cursor string) (dto.Page, error) {
	if actor == "" || d.Store == nil {
		return dto.Page{}, entity.ErrUnavailable
	}
	page, err := d.Store.DiscoverableCandidates(ctx, cursor)
	if err != nil {
		return dto.Page{}, err
	}
	items := page.Items[:0]
	for _, item := range page.Items {
		if item.Status == "ACTIVE" {
			items = append(items, item)
		}
	}
	page.Items = items
	return page, nil
}
