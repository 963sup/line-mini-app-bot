package query

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/port"
)

func Detail(ctx context.Context, store port.Store, actor, id string) (*dto.Detail, error) {
	return store.Enterprise(ctx, actor, id)
}
