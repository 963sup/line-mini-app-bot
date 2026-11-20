package port

import (
	"context"
	"github.com/963sup/line-mini-app-bot/internal/business/journal/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
)

// Operations is implemented by the membership-aware persistent service.
// Actor is always supplied by the authenticated transport. Implementations
// authorize each read, redact inaccessible references and scope create retries
// by actor and key before returning data.
type Operations interface {
	Create(ctx context.Context, actor, key string, draft entity.Draft) (entity.Entry, error)
	Get(ctx context.Context, actor, id string) (entity.Entry, error)
	List(ctx context.Context, actor string, filter dto.Filter) (dto.Page, error)
}
