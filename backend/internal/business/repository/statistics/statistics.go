// Package statistics provides a read-only, complete repository summary.
package statistics

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrNotFound    = errors.New("repository not found")
	ErrUnavailable = errors.New("statistics unavailable")
)

type (
	Statistics struct {
		OpenIssues, InProgressIssues, CompletedIssues, Discussions, Milestones, Projects int64
		ObservedAt                                                                       time.Time
	}
	Repositories interface {
		Get(context.Context, string, string) (repository.Repository, error)
	}
	Store interface {
		Statistics(context.Context, repository.Repository) (Statistics, error)
	}
	Service struct {
		Repositories Repositories
		Store        Store
	}
)

func (s *Service) Get(ctx context.Context, actor, id string) (Statistics, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return Statistics{}, ErrUnavailable
	}
	r, e := s.Repositories.Get(ctx, actor, id)
	if errors.Is(e, repository.ErrNotFound) {
		return Statistics{}, ErrNotFound
	}
	if e != nil {
		return Statistics{}, e
	}
	return s.Store.Statistics(ctx, r)
}
