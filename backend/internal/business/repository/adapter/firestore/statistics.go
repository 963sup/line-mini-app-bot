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
	if !s.ready() {
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
		collection := "users"
		if r.OwnerType == repositoryDomainEntity.OwnerOrganization {
			collection = "organizations"
		}
		owner := s.channel().Collection(collection).Doc(r.OwnerID)
		for _, kind := range []string{"issues", "discussions", "milestones", "projects"} {
			query := parent.Collection(kind).Query
			if kind == "projects" {
				query = owner.Collection("projects").Where("defaultRepositoryId", "==", r.ID)
			}
			snaps, e := tx.Documents(query).GetAll()
			if e != nil {
				return e
			}
			for _, snap := range snaps {
				var d struct {
					SchemaVersion       int64  `firestore:"schemaVersion"`
					ID                  string `firestore:"id"`
					RepositoryID        string `firestore:"repositoryId"`
					OwnerID             string `firestore:"ownerId"`
					DefaultRepositoryID string `firestore:"defaultRepositoryId"`
					State               string `firestore:"state"`
				}
				if e := snap.DataTo(&d); e != nil {
					return e
				}
				if d.SchemaVersion != 1 || d.ID != snap.Ref.ID {
					return repositoryDomainEntityStatistics.ErrUnavailable
				}
				if kind == "projects" {
					if d.OwnerID != r.OwnerID || d.DefaultRepositoryID != r.ID {
						return repositoryDomainEntityStatistics.ErrUnavailable
					}
				} else if d.RepositoryID != r.ID {
					return repositoryDomainEntityStatistics.ErrUnavailable
				}
				switch kind {
				case "issues":
					switch d.State {
					case "open":
						result.OpenIssues++
					case "in_progress":
						result.InProgressIssues++
					case "completed":
						result.CompletedIssues++
					default:
						return repositoryDomainEntityStatistics.ErrUnavailable
					}
				case "discussions":
					result.Discussions++
				case "milestones":
					result.Milestones++
				case "projects":
					result.Projects++
				}
			}
		}
		return nil
	}, db.ReadOnly)
	if err != nil {
		return repositoryDomainEntityStatistics.Statistics{}, err
	}
	result.ObservedAt = time.Now().UTC()
	return result, nil
}
