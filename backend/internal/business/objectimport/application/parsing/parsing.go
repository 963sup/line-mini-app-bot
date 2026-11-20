package parsing

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"errors"

	objectimportApplicationPortParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/parsing"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/parsing"
)

type Service struct {
	Repositories objectimportApplicationPortParsing.Repositories
	Producer     objectimportApplicationPortParsing.Producer
}

func (s *Service) Parse(ctx context.Context, actor, rid string, input objectimportDomainEntityParsing.Input) (objectimportDomainEntityParsing.Result, error) {
	if !input.Consent {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntity.ErrInput
	}
	if s == nil || s.Repositories == nil {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntity.ErrUnavailable
	}
	if actor == "" || !repositoryDomainEntity.ValidID(rid) {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntity.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repositoryDomainEntity.ErrNotFound) {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntity.ErrNotFound
	}
	if e != nil {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntity.ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repositoryDomainEntity.ValidOwner(r.OwnerID) {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntity.ErrNotFound
	}
	source, _, e := objectimportDomainEntity.ValidateSource(input.Source)
	if e != nil {
		return objectimportDomainEntityParsing.Result{}, e
	}
	if s.Producer == nil {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntityParsing.ErrNotConfigured
	}
	rows, e := s.Producer.Candidates(ctx, source)
	if e != nil {
		return objectimportDomainEntityParsing.Result{}, e
	}
	j, _, e := objectimportDomainEntity.Prepare(objectimportDomainEntity.Creation{SourceDocument: source.SourceDocument, SourceRevision: source.SourceRevision, Pages: source.Pages, Rows: rows})
	if e != nil {
		return objectimportDomainEntityParsing.Result{}, objectimportDomainEntityParsing.ErrOutput
	}
	result := objectimportDomainEntityParsing.Result{Rows: make([]objectimportDomainEntity.Candidate, 0, len(j.Rows)), SourceHash: j.SourceHash}
	for _, r := range j.Rows {
		result.Rows = append(result.Rows, objectimportDomainEntity.Candidate{Fields: r.Fields, SourceItemNo: r.SourceItemNo, Evidence: r.Evidence})
	}
	return result, nil
}
