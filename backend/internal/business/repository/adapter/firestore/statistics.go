package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/milestones"
	repositoryDomainEntityStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/statistics"
)

import (
	"context"
	"errors"
	"time"

	db "cloud.google.com/go/firestore"
)

func (s *Store) Statistics(ctx context.Context, r repositoryDomainEntity.Repository) (result repositoryDomainEntityStatistics.Statistics, err error) {
	if !s.ready() || s.CountIssues == nil || s.CountDiscussions == nil || s.CountProjects == nil {
		return result, repositoryDomainEntityStatistics.ErrUnavailable
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = repositoryDomainEntityStatistics.Statistics{}
		if e := s.checkActiveRepository(tx, r); e != nil {
			if errors.Is(e, repositoryDomainEntityMilestones.ErrNotFound) {
				return repositoryDomainEntityStatistics.ErrNotFound
			}
			return e
		}
		parent := s.repositoryRef(r.ID)
		issues, err := s.CountIssues(func(q db.Query) *db.DocumentIterator { return tx.Documents(q) }, parent, "")
		if err != nil {
			return err
		}
		result.OpenIssues, result.InProgressIssues, result.CompletedIssues = issues.Open, issues.InProgress, issues.Completed
		result.Discussions, err = s.CountDiscussions(tx, parent)
		if err != nil {
			return err
		}
		result.Projects, err = s.CountProjects(tx, r.OwnerID, r.ID)
		if err != nil {
			return err
		}
		rows, err := tx.Documents(parent.Collection("milestones").Select("id", "repositoryId", "schemaVersion")).GetAll()
		if err != nil {
			return err
		}
		for _, row := range rows {
			var doc milestoneDocument
			if err := row.DataTo(&doc); err != nil {
				return err
			}
			if doc.SchemaVersion != 1 || row.Ref.ID != doc.ID || doc.RepositoryID != r.ID {
				return repositoryDomainEntityStatistics.ErrUnavailable
			}
		}
		result.Milestones = int64(len(rows))
		return nil
	}, db.ReadOnly)
	if err != nil {
		return repositoryDomainEntityStatistics.Statistics{}, err
	}
	result.ObservedAt = time.Now().UTC()
	return result, nil
}
