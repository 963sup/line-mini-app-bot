// Package automatic coordinates immutable model snapshots and resumable Object creation.
package automatic

import (
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

import (
	"context"
	"encoding/hex"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrOutput       = errors.New("automatic model output rejected")
	ErrTypeRequired = errors.New("select a batch type for a new import intent")
)

type Candidate struct {
	Name, Quantity, Unit, SourceItemNo, Type, Quote string
	Pages                                           []int
}
type Row struct {
	feature.Fields
	ID, SourceItemNo, SourcePages, ResultObjectID string
}
type Run struct {
	ID, RepositoryID, CreatorID, IntentHash    string
	SourceDocument, SourceRevision, SourceHash string
	Status, Reason                             string
	Rows                                       []Row
	CreatedAt, UpdatedAt                       time.Time
}
type Producer interface {
	Rows(context.Context, feature.Source) ([]Candidate, error)
}

// Update must load/check scope and ownership, then atomically persist the callback result.
// A nil current run represents absence; only the initial intent operation accepts absence.
type Store interface {
	Update(context.Context, repository.Repository, string, string, func(*Run) (Run, error)) (Run, error)
}
type Service struct {
	Repositories feature.Repositories
	Objects      feature.Objects
	Store        Store
	Producer     Producer
}

func (s *Service) scope(ctx context.Context, actor, rid string) (repository.Repository, error) {
	if s == nil || s.Repositories == nil || s.Store == nil {
		return repository.Repository{}, feature.ErrUnavailable
	}
	if actor == "" || !repository.ValidID(rid) {
		return repository.Repository{}, feature.ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repository.ErrNotFound) {
		return repository.Repository{}, feature.ErrNotFound
	}
	if e != nil {
		return repository.Repository{}, feature.ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repository.ValidOwner(r.OwnerID) {
		return repository.Repository{}, feature.ErrNotFound
	}
	return r, nil
}

func (s *Service) update(ctx context.Context, actor, rid, id string, apply func(*Run) (Run, error)) (Run, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Run{}, e
	}
	return s.Store.Update(ctx, r, actor, id, apply)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (Run, error) {
	if !repository.ValidID(id) {
		return Run{}, feature.ErrNotFound
	}
	return s.update(ctx, actor, rid, id, func(r *Run) (Run, error) {
		if r == nil {
			return Run{}, feature.ErrNotFound
		}
		return *r, nil
	})
}

func content(j Run, r Row) objectDomainEntity.Content {
	return objectDomainEntity.Content{Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceItemNo: r.SourceItemNo, SourcePages: r.SourcePages}
}

// indexItems binds quotes to the physical item column identified by the PDF reader.
func indexItems(source feature.Source) map[string]map[int]string {
	items := map[string]map[int]string{}
	id := ""
	for _, page := range source.Pages {
		for _, line := range strings.Split(page.Text, "\n") {
			if rest, ok := strings.CutPrefix(line, "[[item:"); ok {
				key, body, ok := strings.Cut(rest, "]] ")
				if !ok || key == "" {
					return nil
				}
				id = key
				line = body
				if items[id] == nil {
					items[id] = map[int]string{}
				}
			}
			if id != "" {
				items[id][page.Number] += line + "\n"
			}
		}
	}
	return items
}

// Prepare accepts page references only, never literal quotations or model authority.
func Prepare(j Run, source feature.Source, candidates []Candidate, selectedType string) ([]Row, error) {
	if selectedType != "per_item" && selectedType != objectDomainEntity.TypeWork && selectedType != objectDomainEntity.TypeNonWork {
		return nil, ErrTypeRequired
	}
	if len(candidates) < 1 || len(candidates) > 100 {
		return nil, feature.ErrInput
	}
	pages := map[int]bool{}
	for _, p := range source.Pages {
		pages[p.Number] = true
	}
	rows := make([]Row, 0, len(candidates))
	seen := map[string]bool{}
	var evidence map[string]map[int]string
	if selectedType == "per_item" {
		evidence = indexItems(source)
		if len(evidence) != len(candidates) {
			return nil, feature.ErrInput
		}
	}
	for i, c := range candidates {
		if len(c.Pages) < 1 || len(c.Pages) > 4 {
			return nil, feature.ErrInput
		}
		refs := slices.Clone(c.Pages)
		slices.Sort(refs)
		refs = slices.Compact(refs)
		labels := []string{}
		for _, p := range refs {
			if !pages[p] {
				return nil, feature.ErrInput
			}
			labels = append(labels, strconv.Itoa(p))
		}
		row := Row{Fields: feature.Fields{Name: c.Name, Quantity: c.Quantity, Unit: c.Unit, Type: selectedType}, ID: feature.Hash([]string{"automatic-row-v1", j.ID, strconv.Itoa(i)})[:32], SourceItemNo: c.SourceItemNo, SourcePages: strings.Join(labels, ",")}
		if selectedType == "per_item" {
			item, ok := evidence[strings.TrimSpace(c.SourceItemNo)]
			if !ok {
				return nil, feature.ErrInput
			}
			row.Type = c.Type
			if row.Type != objectDomainEntity.TypeWork && row.Type != objectDomainEntity.TypeNonWork && row.Type != "needs_review" {
				return nil, feature.ErrInput
			}
			row.Description = strings.TrimSpace(c.Quote)
			matched := false
			for _, ref := range refs {
				if row.Description != "" && strings.Contains(item[ref], row.Description) {
					matched = true
					break
				}
			}
			if !matched {
				row.Type = "needs_review"
				row.Description = ""
			}
		}
		normalized, e := objectDomainEntity.Normalize(content(j, row))
		if e != nil || normalized.SourceItemNo == "" || seen[normalized.SourceItemNo] {
			return nil, feature.ErrInput
		}
		row.Fields = feature.Fields{Name: normalized.Name, Description: normalized.Description, Type: normalized.Type, Quantity: normalized.Quantity, Unit: normalized.Unit, MaterialCode: normalized.MaterialCode}
		row.SourceItemNo = normalized.SourceItemNo
		seen[row.SourceItemNo] = true
		rows = append(rows, row)
	}
	return rows, nil
}

func stamp(j Run) time.Time {
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !now.After(j.UpdatedAt) {
		return j.UpdatedAt.Add(time.Microsecond)
	}
	return now
}

func (s *Service) reason(ctx context.Context, actor, rid string, j Run, reason string) (Run, error) {
	return s.update(ctx, actor, rid, j.ID, func(current *Run) (Run, error) {
		if current == nil {
			return Run{}, feature.ErrNotFound
		}
		out := *current
		// A stale parser failure must never overwrite a winner's frozen snapshot.
		if out.Status == j.Status && out.Status != "completed" {
			out.Reason = reason
			out.UpdatedAt = stamp(out)
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

func (s *Service) Create(ctx context.Context, actor, rid, key string, input feature.Source, selectedType string) (Run, error) {
	return s.create(ctx, actor, rid, key, input, selectedType, nil)
}

func (s *Service) Classify(ctx context.Context, actor, rid, key string, input feature.Source, decisions map[string]string) (Run, error) {
	return s.create(ctx, actor, rid, key, input, "per_item", decisions)
}

func (s *Service) create(ctx context.Context, actor, rid, key string, input feature.Source, selectedType string, decisions map[string]string) (Run, error) {
	if _, e := s.scope(ctx, actor, rid); e != nil {
		return Run{}, e
	}
	if !validKey(key) {
		return Run{}, feature.ErrInput
	}
	source, digest, e := feature.ValidateSource(input)
	if e != nil {
		return Run{}, e
	}
	if selectedType != "" && selectedType != "per_item" && selectedType != objectDomainEntity.TypeWork && selectedType != objectDomainEntity.TypeNonWork {
		return Run{}, ErrTypeRequired
	}
	id := feature.Hash([]string{"automatic-import-v1", actor, rid, key})[:32]
	intent := feature.Hash([]string{"automatic-intent-v1", source.SourceDocument, source.SourceRevision, digest})
	if selectedType != "" {
		intent = feature.Hash([]string{"automatic-intent-v2", source.SourceDocument, source.SourceRevision, digest, selectedType})
	}
	j, e := s.update(ctx, actor, rid, id, func(current *Run) (Run, error) {
		if current != nil {
			if current.IntentHash != intent {
				return Run{}, feature.ErrCreationConflict
			}
			if current.Status == "pending" && selectedType == "" {
				return Run{}, ErrTypeRequired
			}
			out := *current
			if len(decisions) > 0 {
				out.Rows = slices.Clone(current.Rows)
				for id, t := range decisions {
					if t != objectDomainEntity.TypeWork && t != objectDomainEntity.TypeNonWork {
						return Run{}, feature.ErrInput
					}
					found := false
					for i, row := range out.Rows {
						if row.ID == id {
							if row.Type != "needs_review" && row.Type != t {
								return Run{}, feature.ErrConflict
							}
							out.Rows[i].Type = t
							found = true
							break
						}
					}
					if !found {
						return Run{}, feature.ErrInput
					}
				}
			}
			return out, nil
		}
		if len(decisions) > 0 {
			return Run{}, feature.ErrInput
		}
		if selectedType == "" {
			return Run{}, ErrTypeRequired
		}
		now := time.Now().UTC().Truncate(time.Microsecond)
		return Run{ID: id, RepositoryID: rid, CreatorID: actor, IntentHash: intent, SourceDocument: source.SourceDocument, SourceRevision: source.SourceRevision, SourceHash: digest, Status: "pending", Rows: []Row{}, CreatedAt: now, UpdatedAt: now}, nil
	})
	if e != nil {
		return Run{}, e
	}
	if j.Status == "pending" {
		if s.Producer == nil {
			return s.reason(ctx, actor, rid, j, "model_unavailable")
		}
		produce := s.Producer.Rows
		if selectedType == "per_item" {
			classifier, ok := s.Producer.(interface {
				ClassifiedRows(context.Context, feature.Source) ([]Candidate, error)
			})
			if !ok {
				return s.reason(ctx, actor, rid, j, "model_unavailable")
			}
			produce = classifier.ClassifiedRows
		}
		candidates, err := produce(ctx, source)
		if err != nil {
			if errors.Is(err, ErrOutput) {
				return s.reason(ctx, actor, rid, j, "output_rejected")
			}
			return s.reason(ctx, actor, rid, j, "model_unavailable")
		}
		rows, err := Prepare(j, source, candidates, selectedType)
		if err != nil {
			return s.reason(ctx, actor, rid, j, "output_rejected")
		}
		j, e = s.update(ctx, actor, rid, id, func(current *Run) (Run, error) {
			if current == nil {
				return Run{}, feature.ErrNotFound
			}
			out := *current
			if out.Status == "pending" {
				out.Rows = rows
				out.Status = "processing"
				out.Reason = ""
				out.UpdatedAt = stamp(out)
			}
			return out, nil
		})
		if e != nil {
			if errors.Is(e, feature.ErrInput) {
				return s.reason(ctx, actor, rid, Run{ID: id, Status: "pending"}, "output_rejected")
			}
			return Run{}, e
		}
	}
	for _, row := range j.Rows {
		if row.ResultObjectID != "" || row.Type == "needs_review" {
			continue
		}
		if s.Objects == nil {
			return s.reason(ctx, actor, rid, j, "creation_unavailable")
		}
		created, err := s.Objects.Create(ctx, actor, rid, feature.Hash([]string{"automatic-object-v1", j.ID, row.ID}), content(j, row))
		if err != nil || !repository.ValidID(created.ID) {
			return s.reason(ctx, actor, rid, j, "creation_unavailable")
		}
		j, e = s.update(ctx, actor, rid, id, func(current *Run) (Run, error) {
			if current == nil {
				return Run{}, feature.ErrNotFound
			}
			out := *current
			out.Rows = slices.Clone(out.Rows)
			found := false
			all := true
			for i, r := range out.Rows {
				if r.ID == row.ID {
					found = true
					if r.ResultObjectID != "" && r.ResultObjectID != created.ID {
						return Run{}, feature.ErrConflict
					}
					out.Rows[i].ResultObjectID = created.ID
				}
				all = all && out.Rows[i].ResultObjectID != ""
			}
			if !found {
				return Run{}, feature.ErrConflict
			}
			out.Reason = ""
			out.UpdatedAt = stamp(out)
			if all {
				out.Status = "completed"
			}
			return out, nil
		})
		if e != nil {
			return Run{}, e
		}
	}
	return j, nil
}

func ValidateStored(j Run) error {
	for _, h := range []string{j.IntentHash, j.SourceHash} {
		if _, e := hex.DecodeString(h); e != nil || len(h) != 64 || strings.ToLower(h) != h {
			return feature.ErrUnavailable
		}
	}
	source, _, err := feature.ValidateSource(feature.Source{SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, Pages: []feature.Page{{Number: 1}}})
	if err != nil || source.SourceDocument != j.SourceDocument || source.SourceRevision != j.SourceRevision {
		return feature.ErrUnavailable
	}
	if !repository.ValidID(j.ID) || !repository.ValidID(j.RepositoryID) || j.CreatorID == "" || len(j.IntentHash) != 64 || len(j.SourceHash) != 64 || j.CreatedAt.IsZero() || j.UpdatedAt.Before(j.CreatedAt) {
		return feature.ErrUnavailable
	}
	if j.Reason != "" && j.Reason != "model_unavailable" && j.Reason != "output_rejected" && j.Reason != "creation_unavailable" {
		return feature.ErrUnavailable
	}
	if j.Status == "pending" {
		if len(j.Rows) != 0 {
			return feature.ErrUnavailable
		}
		return nil
	}
	if j.Status != "processing" && j.Status != "completed" || len(j.Rows) < 1 || len(j.Rows) > 100 {
		return feature.ErrUnavailable
	}
	ids := map[string]bool{}
	items := map[string]bool{}
	all := true
	for _, r := range j.Rows {
		n, e := objectDomainEntity.Normalize(content(j, r))
		if e != nil || n != content(j, r) || !repository.ValidID(r.ID) || ids[r.ID] || items[r.SourceItemNo] || r.SourceItemNo == "" || r.SourcePages == "" {
			return feature.ErrUnavailable
		}
		ids[r.ID] = true
		items[r.SourceItemNo] = true
		if r.ResultObjectID != "" && !repository.ValidID(r.ResultObjectID) {
			return feature.ErrUnavailable
		}
		all = all && r.ResultObjectID != ""
	}
	if (j.Status == "completed") != all || j.Status == "completed" && j.Reason != "" {
		return feature.ErrUnavailable
	}
	return nil
}
