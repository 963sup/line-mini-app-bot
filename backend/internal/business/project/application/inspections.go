package application

import (
	"context"
	"strings"
	"unicode/utf8"

	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

func (s *Service) Inspections(ctx context.Context, actor, ownerID, projectID, itemID string, note *string) ([]projectDomainEntity.InspectionSubmission, error) {
	if !projectDomainEntity.ValidID(itemID) {
		return nil, projectDomainEntity.ErrInput
	}
	var text string
	if note != nil {
		text = strings.TrimSpace(*note)
		if text == "" || !utf8.ValidString(text) || utf8.RuneCountInString(text) > 2000 {
			return nil, projectDomainEntity.ErrInput
		}
	}
	scope, _, err := s.projectScope(ctx, actor, ownerID, projectID)
	if err != nil {
		return nil, err
	}
	return s.Store.Inspections(ctx, scope, projectID, itemID, text)
}
