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
	pages := map[int]string{}
	textOnly := false
	for _, p := range source.Pages {
		pages[p.Number] = p.Text
		textOnly = textOnly || p.PDFTextOnly
	}
	rows := make([]Row, 0, len(candidates))
	seen := map[string]bool{}
	var evidence map[string]map[int]string
	pdfItems := map[string]objectimportDomainEntity.PDFItem{}
	pdf := false
	for _, page := range source.Pages {
		pdf = pdf || page.PDFItems != nil
		for _, item := range page.PDFItems {
			if _, duplicate := pdfItems[item.ItemNo]; duplicate {
				return nil, objectimportDomainEntity.ErrInput
			}
			pdfItems[item.ItemNo] = item
		}
	}
	if selectedType == "per_item" && !textOnly {
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
			if _, ok := pages[p]; !ok {
				return nil, objectimportDomainEntity.ErrInput
			}
			labels = append(labels, strconv.Itoa(p))
		}
		row := Row{Fields: objectimportDomainEntity.Fields{Name: c.Name, Quantity: c.Quantity, Unit: c.Unit, Type: selectedType}, ID: objectimportDomainEntity.Hash([]string{"automatic-row-v1", j.ID, strconv.Itoa(i)})[:32], SourceItemNo: c.SourceItemNo, SourcePages: strings.Join(labels, ",")}
		if textOnly {
			row.Type = "needs_review"
			row.Description = strings.Join(strings.Fields(c.Quote), " ")
			matched := false
			for _, ref := range refs {
				if row.Description != "" && strings.Contains(strings.Join(strings.Fields(pages[ref]), " "), row.Description) {
					matched = true
					break
				}
			}
			if !matched {
				return nil, objectimportDomainEntity.ErrInput
			}
		} else if selectedType == "per_item" {
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
		if pdf && selectedType == "per_item" {
			original, ok := pdfItems[strings.TrimSpace(c.SourceItemNo)]
			if !ok {
				return nil, objectimportDomainEntity.ErrInput
			}
			row.Quantity, row.Unit = original.Quantity, original.Unit
			if _, err := objectDomainEntity.Normalize(ObjectContent(j, row)); err != nil {
				row.Type = "needs_review"
			}
			var itemText strings.Builder
			for _, text := range evidence[original.ItemNo] {
				itemText.WriteString(text)
				itemText.WriteByte(' ')
			}
			if !supportsPDFType(row.Type, row.Description) || !supportsPDFType(row.Type, itemText.String()) {
				row.Type = "needs_review"
			}
			if row.Type == "needs_review" && row.Description == "" {
				row.Description = strings.TrimSpace(original.Name + " " + original.Quantity + " " + original.Unit)
			}
		}

		normalized, e := normalizeRow(j, row)
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

// An unresolved PDF row can retain unreadable source values for manual correction.
func normalizeRow(run Run, row Row) (objectDomainEntity.Content, error) {
	content := ObjectContent(run, row)
	if row.Type != "needs_review" || !strings.HasPrefix(run.SourceDocument, "pdf:") {
		return objectDomainEntity.Normalize(content)
	}
	probe := content
	probe.Quantity = "1"
	if strings.TrimSpace(probe.Unit) == "" {
		probe.Unit = "?"
	}
	normalized, err := objectDomainEntity.Normalize(probe)
	if err != nil {
		return normalized, err
	}
	quantityProbe := probe
	quantityProbe.Quantity = content.Quantity
	if quantity, err := objectDomainEntity.Normalize(quantityProbe); err == nil {
		normalized.Quantity = quantity.Quantity
	} else {
		normalized.Quantity = ""
	}
	normalized.Unit = strings.TrimSpace(content.Unit)
	return normalized, nil
}

// Automatic creation requires an affirmative, unambiguous phrase in a verified quote.
func supportsPDFType(kind, quote string) bool {
	quote = strings.ToLower(quote)
	supply := containsAny(quote, "僅供料", "只供料", "不含安裝", "supply only", "without installation")
	work := containsAny(quote, "供應並安裝", "含安裝", "供料及安裝", "supply and install", "including installation")
	// A negative phrase can itself contain the positive phrase.
	if strings.Contains(quote, "不含安裝") {
		work = containsAny(strings.ReplaceAll(quote, "不含安裝", ""), "供應並安裝", "含安裝", "供料及安裝", "supply and install", "including installation")
	}
	return kind == objectDomainEntity.TypeWork && work && !supply ||
		kind == objectDomainEntity.TypeNonWork && supply && !work
}

func containsAny(text string, phrases ...string) bool {
	for _, phrase := range phrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
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
		n, e := normalizeRow(j, r)
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
