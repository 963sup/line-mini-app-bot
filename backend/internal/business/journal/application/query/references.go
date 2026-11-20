package query

import (
	"context"
	"errors"
	"strings"

	group "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	groupentity "github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/journal/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"
	project "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	projectentity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repoentity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

// ReferenceCatalog lists repositories before projects. Each source owns its
// cursor validation and rechecks the actor's access on every page.
type ReferenceCatalog struct {
	Groups       *group.Service
	Repositories *repository.Service
	Projects     *project.Service
}

func (c ReferenceCatalog) ListReferenceTargets(ctx context.Context, actor, groupID, cursor string) (dto.TargetPage, error) {
	var result dto.TargetPage
	if _, err := c.Groups.Get(ctx, actor, groupID); err != nil {
		if errors.Is(err, groupentity.ErrNotFound) {
			return result, entity.ErrNotFound
		}
		return result, err
	}
	if cursor == "" || strings.HasPrefix(cursor, "r:") {
		page, err := c.Repositories.List(ctx, actor, strings.TrimPrefix(cursor, "r:"), groupID)
		if errors.Is(err, repoentity.ErrCursor) {
			return result, entity.ErrInput
		}
		if err != nil {
			return result, err
		}
		for _, r := range page.Items {
			result.Items = append(result.Items, dto.Target{Kind: "repository", ID: r.ID, Name: r.Name})
		}
		result.NextCursor = "p:"
		if page.NextCursor != "" {
			result.NextCursor = "r:" + page.NextCursor
		}
		return result, nil
	}
	after, ok := strings.CutPrefix(cursor, "p:")
	if !ok {
		return result, entity.ErrInput
	}
	page, err := c.Projects.List(ctx, actor, groupID, after)
	if errors.Is(err, projectentity.ErrInput) {
		return result, entity.ErrInput
	}
	if errors.Is(err, projectentity.ErrNotFound) {
		return result, entity.ErrNotFound
	}
	if err != nil {
		return result, err
	}
	for _, p := range page.Items {
		result.Items = append(result.Items, dto.Target{Kind: "project", ID: p.ID, Name: p.Title})
	}
	if page.NextCursor != "" {
		result.NextCursor = "p:" + page.NextCursor
	}
	return result, nil
}
