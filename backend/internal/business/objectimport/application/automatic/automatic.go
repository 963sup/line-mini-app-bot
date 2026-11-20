package automatic

import (
	"context"
	"errors"
	"slices"
	"time"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	objectimportApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port"
	objectimportApplicationPortAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/automatic"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type Service struct {
	Repositories objectimportApplicationPort.Repositories
	Objects      objectimportApplicationPort.Objects
	Store        objectimportApplicationPortAutomatic.Store
	Producer     objectimportApplicationPortAutomatic.Producer
}

func (s *Service) scope(ctx context.Context, actor, rid string) (repository.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repository.Repository{}, objectimportDomainEntity.ErrUnavailable
	}
	if actor == "" || !repository.ValidID(rid) {
		return repository.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repository.ErrNotFound) {
		return repository.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	if e != nil {
		return repository.Repository{}, objectimportDomainEntity.ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repository.ValidOwner(r.OwnerID) {
		return repository.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	return r, nil
}

func (s *Service) update(ctx context.Context, actor, rid, id string, apply func(*objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error)) (objectimportDomainEntityAutomatic.Run, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return objectimportDomainEntityAutomatic.Run{}, e
	}
	return s.Store.Update(ctx, r, actor, id, apply)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (objectimportDomainEntityAutomatic.Run, error) {
	if !repository.ValidID(id) {
		return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrNotFound
	}
	return s.update(ctx, actor, rid, id, func(r *objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error) {
		if r == nil {
			return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrNotFound
		}
		return *r, nil
	})
}

func (s *Service) reason(ctx context.Context, actor, rid string, j objectimportDomainEntityAutomatic.Run, reason string) (objectimportDomainEntityAutomatic.Run, error) {
	return s.update(ctx, actor, rid, j.ID, func(current *objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error) {
		if current == nil {
			return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrNotFound
		}
		out := *current
		// A stale parser failure must never overwrite a winner's frozen snapshot.
		if out.Status == j.Status && out.Status != "completed" {
			out.Reason = reason
			out.UpdatedAt = objectimportDomainEntityAutomatic.Stamp(out)
		}
		return out, nil
	})
}

func validKey(key string) bool {
	if len(key) < 16 || len(key) > 128 {
		return false
	}
	for _, r := range key {
		allowed := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-'
		if !allowed {
			return false
		}
	}
	return true
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, input objectimportDomainEntity.Source, selectedType string) (objectimportDomainEntityAutomatic.Run, error) {
	return s.create(ctx, actor, rid, key, input, selectedType, nil)
}

func (s *Service) Classify(ctx context.Context, actor, rid, key string, input objectimportDomainEntity.Source, decisions map[string]string) (objectimportDomainEntityAutomatic.Run, error) {
	return s.create(ctx, actor, rid, key, input, "per_item", decisions)
}

func (s *Service) create(ctx context.Context, actor, rid, key string, input objectimportDomainEntity.Source, selectedType string, decisions map[string]string) (objectimportDomainEntityAutomatic.Run, error) {
	if _, e := s.scope(ctx, actor, rid); e != nil {
		return objectimportDomainEntityAutomatic.Run{}, e
	}
	if !validKey(key) {
		return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrInput
	}
	source, digest, e := objectimportDomainEntity.ValidateSource(input)
	if e != nil {
		return objectimportDomainEntityAutomatic.Run{}, e
	}
	if selectedType != "" && selectedType != "per_item" && selectedType != objectDomainEntity.TypeWork && selectedType != objectDomainEntity.TypeNonWork {
		return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntityAutomatic.ErrTypeRequired
	}
	id := objectimportDomainEntity.Hash([]string{"automatic-import-v1", actor, rid, key})[:32]
	intent := objectimportDomainEntity.Hash([]string{"automatic-intent-v1", source.SourceDocument, source.SourceRevision, digest})
	if selectedType != "" {
		intent = objectimportDomainEntity.Hash([]string{"automatic-intent-v2", source.SourceDocument, source.SourceRevision, digest, selectedType})
	}
	j, e := s.update(ctx, actor, rid, id, func(current *objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error) {
		if current != nil {
			if current.IntentHash != intent {
				return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrCreationConflict
			}
			if current.Status == "pending" && selectedType == "" {
				return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntityAutomatic.ErrTypeRequired
			}
			out := *current
			if len(decisions) > 0 {
				out.Rows = slices.Clone(current.Rows)
				for id, t := range decisions {
					if t != objectDomainEntity.TypeWork && t != objectDomainEntity.TypeNonWork {
						return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrInput
					}
					found := false
					for i, row := range out.Rows {
						if row.ID == id {
							if row.Type != "needs_review" && row.Type != t {
								return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrConflict
							}
							out.Rows[i].Type = t
							found = true
							break
						}
					}
					if !found {
						return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrInput
					}
				}
			}
			return out, nil
		}
		if len(decisions) > 0 {
			return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrInput
		}
		if selectedType == "" {
			return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntityAutomatic.ErrTypeRequired
		}
		now := time.Now().UTC().Truncate(time.Microsecond)
		return objectimportDomainEntityAutomatic.Run{ID: id, RepositoryID: rid, CreatorID: actor, IntentHash: intent, SourceDocument: source.SourceDocument, SourceRevision: source.SourceRevision, SourceHash: digest, Status: "pending", Rows: []objectimportDomainEntityAutomatic.Row{}, CreatedAt: now, UpdatedAt: now}, nil
	})
	if e != nil {
		return objectimportDomainEntityAutomatic.Run{}, e
	}
	if j.Status == "pending" {
		if s.Producer == nil {
			return s.reason(ctx, actor, rid, j, "model_unavailable")
		}
		produce := s.Producer.Rows
		if selectedType == "per_item" {
			classifier, ok := s.Producer.(interface {
				ClassifiedRows(context.Context, objectimportDomainEntity.Source) ([]objectimportDomainEntityAutomatic.Candidate, error)
			})
			if !ok {
				return s.reason(ctx, actor, rid, j, "model_unavailable")
			}
			produce = classifier.ClassifiedRows
		}
		candidates, err := produce(ctx, source)
		if err != nil {
			if errors.Is(err, objectimportDomainEntityAutomatic.ErrOutput) {
				return s.reason(ctx, actor, rid, j, "output_rejected")
			}
			return s.reason(ctx, actor, rid, j, "model_unavailable")
		}
		rows, err := objectimportDomainEntityAutomatic.Prepare(j, source, candidates, selectedType)
		if err != nil {
			return s.reason(ctx, actor, rid, j, "output_rejected")
		}
		j, e = s.update(ctx, actor, rid, id, func(current *objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error) {
			if current == nil {
				return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrNotFound
			}
			out := *current
			if out.Status == "pending" {
				out.Rows = rows
				out.Status = "processing"
				out.Reason = ""
				out.UpdatedAt = objectimportDomainEntityAutomatic.Stamp(out)
			}
			return out, nil
		})
		if e != nil {
			if errors.Is(e, objectimportDomainEntity.ErrInput) {
				return s.reason(ctx, actor, rid, objectimportDomainEntityAutomatic.Run{ID: id, Status: "pending"}, "output_rejected")
			}
			return objectimportDomainEntityAutomatic.Run{}, e
		}
	}
	for _, row := range j.Rows {
		if row.ResultObjectID != "" || row.Type == "needs_review" {
			continue
		}
		if s.Objects == nil {
			return s.reason(ctx, actor, rid, j, "creation_unavailable")
		}
		created, err := s.Objects.Create(ctx, actor, rid, objectimportDomainEntity.Hash([]string{"automatic-object-v1", j.ID, row.ID}), objectimportDomainEntityAutomatic.ObjectContent(j, row))
		if err != nil || !repository.ValidID(created.ID) {
			return s.reason(ctx, actor, rid, j, "creation_unavailable")
		}
		j, e = s.update(ctx, actor, rid, id, func(current *objectimportDomainEntityAutomatic.Run) (objectimportDomainEntityAutomatic.Run, error) {
			if current == nil {
				return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrNotFound
			}
			out := *current
			out.Rows = slices.Clone(out.Rows)
			found := false
			all := true
			for i, r := range out.Rows {
				if r.ID == row.ID {
					found = true
					if r.ResultObjectID != "" && r.ResultObjectID != created.ID {
						return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrConflict
					}
					out.Rows[i].ResultObjectID = created.ID
				}
				all = all && out.Rows[i].ResultObjectID != ""
			}
			if !found {
				return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrConflict
			}
			out.Reason = ""
			out.UpdatedAt = objectimportDomainEntityAutomatic.Stamp(out)
			if all {
				out.Status = "completed"
			}
			return out, nil
		})
		if e != nil {
			return objectimportDomainEntityAutomatic.Run{}, e
		}
	}
	return j, nil
}
