package automatic

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"encoding/hex"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

var ErrOutput = errors.New("automatic model output rejected")

var ErrTypeRequired = errors.New("select a batch type for a new import intent")

type Candidate struct {
	Name, Quantity, Unit, SourceItemNo, Type, Quote string
	Pages                                           []int
}

type Row struct {
	objectimportDomainEntity.Fields
	ID, SourceItemNo, SourcePages, ResultObjectID string
}

type Run struct {
	ID, RepositoryID, CreatorID, IntentHash    string
	SourceDocument, SourceRevision, SourceHash string
	Status, Reason                             string
	Rows                                       []Row
	CreatedAt, UpdatedAt                       time.Time
}

func ObjectContent(j Run, r Row) objectDomainEntity.Content {
	return objectDomainEntity.Content{Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceItemNo: r.SourceItemNo, SourcePages: r.SourcePages}
}

// indexItems binds quotes to the physical item column identified by the PDF reader.
func indexItems(source objectimportDomainEntity.Source) map[string]map[int]string {
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
func Prepare(j Run, source objectimportDomainEntity.Source, candidates []Candidate, selectedType string) ([]Row, error) {
	if selectedType != "per_item" && selectedType != objectDomainEntity.TypeWork && selectedType != objectDomainEntity.TypeNonWork {
		return nil, ErrTypeRequired
	}
	if len(candidates) < 1 || len(candidates) > 100 {
		return nil, objectimportDomainEntity.ErrInput
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
			return nil, objectimportDomainEntity.ErrInput
		}
	}
	for i, c := range candidates {
		if len(c.Pages) < 1 || len(c.Pages) > 4 {
			return nil, objectimportDomainEntity.ErrInput
		}
		refs := slices.Clone(c.Pages)
		slices.Sort(refs)
		refs = slices.Compact(refs)
		labels := []string{}
		for _, p := range refs {
			if !pages[p] {
				return nil, objectimportDomainEntity.ErrInput
			}
			labels = append(labels, strconv.Itoa(p))
		}
		row := Row{Fields: objectimportDomainEntity.Fields{Name: c.Name, Quantity: c.Quantity, Unit: c.Unit, Type: selectedType}, ID: objectimportDomainEntity.Hash([]string{"automatic-row-v1", j.ID, strconv.Itoa(i)})[:32], SourceItemNo: c.SourceItemNo, SourcePages: strings.Join(labels, ",")}
		if selectedType == "per_item" {
			item, ok := evidence[strings.TrimSpace(c.SourceItemNo)]
			if !ok {
				return nil, objectimportDomainEntity.ErrInput
			}
			row.Type = c.Type
			if row.Type != objectDomainEntity.TypeWork && row.Type != objectDomainEntity.TypeNonWork && row.Type != "needs_review" {
				return nil, objectimportDomainEntity.ErrInput
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
		normalized, e := objectDomainEntity.Normalize(ObjectContent(j, row))
		if e != nil || normalized.SourceItemNo == "" || seen[normalized.SourceItemNo] {
			return nil, objectimportDomainEntity.ErrInput
		}
		row.Fields = objectimportDomainEntity.Fields{Name: normalized.Name, Description: normalized.Description, Type: normalized.Type, Quantity: normalized.Quantity, Unit: normalized.Unit, MaterialCode: normalized.MaterialCode}
		row.SourceItemNo = normalized.SourceItemNo
		seen[row.SourceItemNo] = true
		rows = append(rows, row)
	}
	return rows, nil
}

func Stamp(j Run) time.Time {
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !now.After(j.UpdatedAt) {
		return j.UpdatedAt.Add(time.Microsecond)
	}
	return now
}

func ValidateStored(j Run) error {
	for _, h := range []string{j.IntentHash, j.SourceHash} {
		if _, e := hex.DecodeString(h); e != nil || len(h) != 64 || strings.ToLower(h) != h {
			return objectimportDomainEntity.ErrUnavailable
		}
	}
	source, _, err := objectimportDomainEntity.ValidateSource(objectimportDomainEntity.Source{SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, Pages: []objectimportDomainEntity.Page{{Number: 1}}})
	if err != nil || source.SourceDocument != j.SourceDocument || source.SourceRevision != j.SourceRevision {
		return objectimportDomainEntity.ErrUnavailable
	}
	if !repositoryDomainEntity.ValidID(j.ID) || !repositoryDomainEntity.ValidID(j.RepositoryID) || j.CreatorID == "" || len(j.IntentHash) != 64 || len(j.SourceHash) != 64 || j.CreatedAt.IsZero() || j.UpdatedAt.Before(j.CreatedAt) {
		return objectimportDomainEntity.ErrUnavailable
	}
	if j.Reason != "" && j.Reason != "model_unavailable" && j.Reason != "output_rejected" && j.Reason != "creation_unavailable" {
		return objectimportDomainEntity.ErrUnavailable
	}
	if j.Status == "pending" {
		if len(j.Rows) != 0 {
			return objectimportDomainEntity.ErrUnavailable
		}
		return nil
	}
	if j.Status != "processing" && j.Status != "completed" || len(j.Rows) < 1 || len(j.Rows) > 100 {
		return objectimportDomainEntity.ErrUnavailable
	}
	ids := map[string]bool{}
	items := map[string]bool{}
	all := true
	for _, r := range j.Rows {
		n, e := objectDomainEntity.Normalize(ObjectContent(j, r))
		if e != nil || n != ObjectContent(j, r) || !repositoryDomainEntity.ValidID(r.ID) || ids[r.ID] || items[r.SourceItemNo] || r.SourceItemNo == "" || r.SourcePages == "" {
			return objectimportDomainEntity.ErrUnavailable
		}
		ids[r.ID] = true
		items[r.SourceItemNo] = true
		if r.ResultObjectID != "" && !repositoryDomainEntity.ValidID(r.ResultObjectID) {
			return objectimportDomainEntity.ErrUnavailable
		}
		all = all && r.ResultObjectID != ""
	}
	if (j.Status == "completed") != all || j.Status == "completed" && j.Reason != "" {
		return objectimportDomainEntity.ErrUnavailable
	}
	return nil
}
