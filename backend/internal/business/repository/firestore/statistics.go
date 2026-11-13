package firestore

import (
	"context"
	"errors"
	"time"

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
		for _, kind := range []string{"issues", "discussions", "milestones", "projects"} {
			snaps, e := tx.Documents(parent.Collection(kind)).GetAll()
			if e != nil {
				return e
			}
			for _, snap := range snaps {
				var d struct {
					SchemaVersion int64  `firestore:"schemaVersion"`
					ID            string `firestore:"id"`
					RepositoryID  string `firestore:"repositoryId"`
					State         string `firestore:"state"`
				}
				if e := snap.DataTo(&d); e != nil {
					return e
				}
				if d.SchemaVersion != 1 || d.ID != snap.Ref.ID || d.RepositoryID != r.ID {
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
