package application

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/query"
)

type (
	DraftPreparation = port.DraftPreparation
	TemplatePreparer = port.TemplatePreparer
	TemplateDraft    = dto.TemplateDraft
)

func (s *Service) PrepareTemplateDraft(ctx context.Context, actor string, in DraftPreparation) (TemplateDraft, error) {
	var preparer port.TemplatePreparer
	if s != nil {
		preparer = s.TemplatePreparer
	}
	return query.PrepareTemplateDraft(ctx, preparer, actor, in)
}
