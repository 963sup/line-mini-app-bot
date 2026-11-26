package command

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/issue/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Operations struct{ port.Access }

func (s *Operations) Create(ctx context.Context, actorID, repositoryID, idempotencyKey string, content entity.Content) (entity.Issue, error) {
	repo, err := s.Scope(ctx, actorID, repositoryID)
	if err != nil {
		return entity.Issue{}, err
	}
	content, err = entity.Normalize(content)
	if err != nil || !creationKey.MatchString(idempotencyKey) {
		return entity.Issue{}, entity.ErrInput
	}
	return s.Store.Create(ctx, repo, actorID, idempotencyKey, content)
}

func (s *Operations) Update(ctx context.Context, actor, repositoryID, id string, input entity.Update) (entity.Issue, error) {
	r, err := s.Scope(ctx, actor, repositoryID)
	if err != nil {
		return entity.Issue{}, err
	}
	if !repositoryDomainEntity.ValidID(id) {
		return entity.Issue{}, entity.ErrNotFound
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	return s.Store.RunUpdate(ctx, r, id, func(work port.UpdateWork) (entity.Issue, error) {
		if err := work.CheckRepository(actor); err != nil {
			return entity.Issue{}, err
		}
		current, err := work.Current()
		if err != nil {
			return entity.Issue{}, err
		}
		updated, err := entity.ApplyUpdate(current, actor, input)
		if err != nil || updated.Version == current.Version {
			return updated, err
		}
		if len(updated.LabelIDs) > 0 {
			if err := work.ValidateLabels(updated.LabelIDs); err != nil {
				return entity.Issue{}, invalidReference(err)
			}
		}
		if updated.MilestoneID != "" {
			if err := work.ValidateMilestone(updated.MilestoneID); err != nil {
				return entity.Issue{}, invalidReference(err)
			}
		}
		updated.UpdatedAt = now
		if now.Before(current.UpdatedAt) {
			updated.UpdatedAt = current.UpdatedAt
		}
		return updated, work.Save(actor, current, updated)
	})
}

func invalidReference(err error) error {
	if errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		return entity.ErrInput
	}
	return err
}
