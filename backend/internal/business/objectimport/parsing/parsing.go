// Package parsing produces validated proposals without creating drafts or Objects.
package parsing

import (
	"context"
	"errors"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrNotConfigured = errors.New("object import model not configured")
	ErrModel         = errors.New("object import model unavailable")
	ErrOutput        = errors.New("object import model output rejected")
)

type Input struct {
	Source  feature.Source
	Consent bool
}
type Result struct {
	Rows       []feature.Candidate
	SourceHash string
}
type Producer interface {
	Candidates(context.Context, feature.Source) ([]feature.Candidate, error)
}
type Repositories interface {
	Get(context.Context, string, string) (repository.Repository, error)
}
type Service struct {
	Repositories Repositories
	Producer     Producer
}

func (s *Service) Parse(ctx context.Context, actor, rid string, input Input) (Result, error) {
	if !input.Consent {
		return Result{}, feature.ErrInput
	}
	if s == nil || s.Repositories == nil {
		return Result{}, feature.ErrUnavailable
	}
	if actor == "" || !repository.ValidID(rid) {
		return Result{}, feature.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repository.ErrNotFound) {
		return Result{}, feature.ErrNotFound
	}
	if e != nil {
		return Result{}, feature.ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repository.ValidGroup(r.GroupID) {
		return Result{}, feature.ErrNotFound
	}
	source, _, e := feature.ValidateSource(input.Source)
	if e != nil {
		return Result{}, e
	}
	if s.Producer == nil {
		return Result{}, ErrNotConfigured
	}
	rows, e := s.Producer.Candidates(ctx, source)
	if e != nil {
		return Result{}, e
	}
	j, _, e := feature.Prepare(feature.Creation{SourceDocument: source.SourceDocument, SourceRevision: source.SourceRevision, Pages: source.Pages, Rows: rows})
	if e != nil {
		return Result{}, ErrOutput
	}
	result := Result{Rows: make([]feature.Candidate, 0, len(j.Rows)), SourceHash: j.SourceHash}
	for _, r := range j.Rows {
		result.Rows = append(result.Rows, feature.Candidate{Fields: r.Fields, SourceItemNo: r.SourceItemNo, Evidence: r.Evidence})
	}
	return result, nil
}
