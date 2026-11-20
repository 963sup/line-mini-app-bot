package port

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
)

type Store interface {
	CreateEnterprise(context.Context, string, string) (*dto.Link, error)
	Enterprise(context.Context, string, string) (*dto.Detail, error)
	UpdateEnterprise(context.Context, string, string, string, bool) (*dto.Link, error)
}
