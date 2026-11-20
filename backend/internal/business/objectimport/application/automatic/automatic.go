package automatic

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"

	knowledge "github.com/963sup/line-mini-app-bot/internal/business/knowledge/application/port"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	objectimportApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port"
	objectimportApplicationPortAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/automatic"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
	objectimportDomainEntityAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity/automatic"
)

type Service struct {
	Knowledge    knowledge.Reader
	Sources      objectimportApplicationPortAutomatic.SourceStore
	TextTable    func(string, []byte) ([]objectimportDomainEntityAutomatic.Row, error)
	Workbook     func([]byte) ([]objectimportDomainEntityAutomatic.Row, error)
	Repositories objectimportApplicationPort.Repositories
	Objects      objectimportApplicationPort.Objects
	Store        objectimportApplicationPortAutomatic.Store
	Producer     objectimportApplicationPortAutomatic.Producer
}

func (s *Service) scope(ctx context.Context, actor, rid string) (repositoryDomainEntity.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repositoryDomainEntity.Repository{}, objectimportDomainEntity.ErrUnavailable
	}
	if actor == "" || !repositoryDomainEntity.ValidID(rid) {
		return repositoryDomainEntity.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repositoryDomainEntity.ErrNotFound) {
		return repositoryDomainEntity.Repository{}, objectimportDomainEntity.ErrNotFound
	}
	if e != nil {
		return repositoryDomainEntity.Repository{}, objectimportDomainEntity.ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repositoryDomainEntity.ValidOwner(r.OwnerID) {
		return repositoryDomainEntity.Repository{}, objectimportDomainEntity.ErrNotFound
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
	if !repositoryDomainEntity.ValidID(id) {
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
	return s.run(ctx, actor, rid, key, source, digest, selectedType, decisions, nil, nil)
}

// ImportTable validates the file after repository authorization. Its rows are
// deterministic; only the user's work/non-work classification remains pending.
func (s *Service) ImportTable(ctx context.Context, actor, rid, key, document string, file []byte, decisions map[string]string) (objectimportDomainEntityAutomatic.Run, error) {
	if _, err := s.scope(ctx, actor, rid); err != nil {
		return objectimportDomainEntityAutomatic.Run{}, err
	}
	sum := sha256.Sum256(file)
	digest := hex.EncodeToString(sum[:])
	format, _, _ := strings.Cut(document, ":")
	if !validKey(key) || len(file) == 0 || len(file) > 300*1024 || document != format+":"+digest {
		return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrInput
	}
	var rows []objectimportDomainEntityAutomatic.Row
	var err error
	switch format {
	case "xlsx":
		if s.Workbook == nil {
			return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrUnavailable
		}
		rows, err = s.Workbook(file)
	case "csv", "tsv", "txt", "md":
		if s.TextTable == nil {
			return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrUnavailable
		}
		rows, err = s.TextTable(format, file)
		file = nil // Text sources are not archived.
	default:
		return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrInput
	}
	if err != nil {
		return objectimportDomainEntityAutomatic.Run{}, err
	}
	return s.run(ctx, actor, rid, key, objectimportDomainEntity.Source{SourceDocument: document}, digest, "per_item", decisions, rows, file)
}

func (s *Service) run(ctx context.Context, actor, rid, key string, source objectimportDomainEntity.Source, digest, selectedType string, decisions map[string]string, structured []objectimportDomainEntityAutomatic.Row, file []byte) (objectimportDomainEntityAutomatic.Run, error) {
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
		if file != nil && s.Sources != nil {
			if err := s.Sources.Put(ctx, rid, actor, j.SourceHash, file); err != nil {
				return objectimportDomainEntityAutomatic.Run{}, objectimportDomainEntity.ErrUnavailable
			}
		}

		rows := structured
		if rows == nil {
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
			rows, err = objectimportDomainEntityAutomatic.Prepare(j, source, candidates, selectedType)
			if err != nil {
				return s.reason(ctx, actor, rid, j, "output_rejected")
			}
		} else {
			for i := range rows {
				rows[i].ID = objectimportDomainEntity.Hash([]string{"workbook-row-v1", j.ID, rows[i].SourceItemNo})[:32]
			}
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
		if err != nil || !repositoryDomainEntity.ValidID(created.ID) {
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

// ReadSourceCell rechecks current run ownership before resolving a frozen source location.
func (s *Service) ReadSourceCell(ctx context.Context, actor, repository, runID, location string) (string, error) {
	j, err := s.Get(ctx, actor, repository, runID)
	if err != nil {
		return "", err
	}
	if s.Sources == nil || s.Workbook == nil {
		return "", objectimportDomainEntity.ErrUnavailable
	}
	if j.CreatorID != actor || j.RepositoryID != repository || j.SourceDocument != "xlsx:"+j.SourceHash {
		return "", objectimportDomainEntity.ErrNotFound
	}
	for _, row := range j.Rows {
		if row.SourcePages == location {
			data, err := s.Sources.Read(ctx, repository, actor, j.SourceHash)
			if err != nil {
				return "", err
			}
			rows, err := s.Workbook(data)
			if err != nil {
				return "", err
			}
			for _, source := range rows {
				if source.SourcePages == location && source.SourceItemNo == row.SourceItemNo {
					return source.Quantity, nil
				}
			}
			break
		}
	}
	return "", objectimportDomainEntity.ErrNotFound
}
