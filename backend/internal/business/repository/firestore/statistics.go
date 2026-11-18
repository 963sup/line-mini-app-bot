package firestore

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/milestones"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/statistics"
)

type statisticsFailure struct{ cause error }

func (e *statisticsFailure) Error() string { return "statistics storage unavailable" }
func (e *statisticsFailure) Unwrap() error { return statistics.ErrUnavailable }
func (s *Store) Statistics(ctx context.Context, r repository.Repository) (result statistics.Statistics, err error) {
	defer func() {
		if err != nil && !errors.Is(err, statistics.ErrNotFound) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			err = &statisticsFailure{err}
		}
	}()
	if !s.planningReady() {
		return result, statistics.ErrUnavailable
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = statistics.Statistics{}
		if e := s.planningContainer(tx, r); e != nil {
			if errors.Is(e, milestones.ErrNotFound) {
				return statistics.ErrNotFound
			}
			return e
		}
		parent := s.planningParent(r.ID)
		personal := identity.IsUserOwnerID(r.OwnerID)
		collection := "groups"
		if personal {
			collection = "users"
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
			if kind == "projects" && !personal && len(snaps) > 0 {
				snapshot, err := tx.Get(owner)
				if err != nil {
					return err
				}
				var space struct{ ID, GroupID, Status string }
				if err = snapshot.DataTo(&space); err != nil {
					return err
				}
				if space.ID != r.OwnerID || space.GroupID != r.OwnerID || space.Status != "ACTIVE" {
					return statistics.ErrUnavailable
				}
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
					return statistics.ErrUnavailable
				}
				if kind == "projects" {
					if d.OwnerID != r.OwnerID || d.DefaultRepositoryID != r.ID {
						return statistics.ErrUnavailable
					}
				} else if d.RepositoryID != r.ID {
					return statistics.ErrUnavailable
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
						return statistics.ErrUnavailable
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
		return statistics.Statistics{}, err
	}
	result.ObservedAt = time.Now().UTC()
	return result, nil
}
