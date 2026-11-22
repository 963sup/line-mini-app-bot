package group

import (
	"context"
	"errors"

	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	groupmodel "github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
	projectmodel "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

type Access struct {
	Groups *groupquery.Service
}

func (a Access) AuthorizeGroup(ctx context.Context, actor, ownerID string) (string, error) {
	if a.Groups == nil {
		return "", projectmodel.ErrUnavailable
	}
	group, err := a.Groups.Get(ctx, actor, ownerID)
	if errors.Is(err, groupmodel.ErrNotFound) {
		return "", projectmodel.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return group.ID, nil
}
