package entity

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

var ErrInput = errors.New("invalid object import")

var ErrNotFound = errors.New("object import not found")

var ErrForbidden = errors.New("object import owner required")

var ErrConflict = errors.New("object import version or state conflict")

var ErrCreationConflict = errors.New("object import creation key conflict")

var ErrUnavailable = errors.New("object import unavailable")

var hashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

const MaxDraftBytes = 256 * 1024

// Fields are editable proposals. Source item identity and evidence remain separate.
type Fields struct {
	Name         string
	Description  string
	Type         string
	Quantity     string
	Unit         string
	MaterialCode string
}

type Evidence struct {
	Page  int
	Quote string
}

type PDFItem struct {
	ItemNo, Name, Quantity, Unit string
}

type Page struct {
	Number int
	Text   string
}

// PDFInput is the server reader output consumed by the automatic importer.
type PDFInput struct {
	Pages    []Page
	Evidence PDFEvidence
}

// PDFEvidence describes what the server independently recovered from a PDF.
type PDFEvidence interface{ pdfEvidence() }

type LocatedPDFItems struct{ Items []PDFItem }
type PDFPageText struct{}

func (LocatedPDFItems) pdfEvidence() {}
func (PDFPageText) pdfEvidence()     {}

type Candidate struct {
	Fields
	SourceItemNo string
	Evidence     []Evidence
}

type Creation struct {
	SourceDocument string
	SourceRevision string
	Pages          []Page
	Rows           []Candidate
}

type Row struct {
	Fields
	ID             string
	SourceItemNo   string
	Evidence       []Evidence
	ResultObjectID string
}

type Job struct {
	ID                         string
	RepositoryID               string
	CreatorID                  string
	Status                     string
	Version                    int64
	ConfirmedFromReviewVersion int64
	SourceDocument             string
	SourceRevision             string
	SourceHash                 string
	Rows                       []Row
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	ReviewedAt                 *time.Time
}

type ReviewedRow struct {
	Fields
	ID string
}

type Review struct {
	ExpectedVersion int64
	Rows            []ReviewedRow
}

func Hash(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func NewID() (string, error) {
	var b [16]byte
	_, e := rand.Read(b[:])
	return hex.EncodeToString(b[:]), e
}

func ObjectContent(j Job, r Row) objectDomainEntity.Content {
	pages := []int{}
	for _, e := range r.Evidence {
		if !slices.Contains(pages, e.Page) {
			pages = append(pages, e.Page)
		}
	}
	slices.Sort(pages)
	labels := []string{}
	for _, p := range pages {
		labels = append(labels, strconv.Itoa(p))
	}
	return objectDomainEntity.Content{Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceItemNo: r.SourceItemNo, SourcePages: strings.Join(labels, ",")}
}

func normalize(j Job, r Row) (Row, error) {
	c, e := objectDomainEntity.Normalize(ObjectContent(j, r))
	if e != nil {
		return Row{}, ErrInput
	}
	r.Fields = Fields{c.Name, c.Description, c.Type, c.Quantity, c.Unit, c.MaterialCode}
	r.SourceItemNo = c.SourceItemNo
	return r, nil
}

// Budget a fixed final snapshot, so later review/result metadata consumes reserved
// capacity instead of reapplying a moving threshold to the same accepted draft.
func ProjectedBytes(j Job) (int, error) {
	j.Rows = append([]Row{}, j.Rows...)
	for i := range j.Rows {
		j.Rows[i].ID = strings.Repeat("f", 32)
		j.Rows[i].ResultObjectID = strings.Repeat("f", 32)
	}
	j.ID = strings.Repeat("f", 32)
	j.RepositoryID = strings.Repeat("f", 32)
	j.Status = "completed"
	j.Version = repositoryDomainEntity.MaxResourceNumber
	j.ConfirmedFromReviewVersion = repositoryDomainEntity.MaxResourceNumber
	stamp := time.Date(2026, 1, 1, 0, 0, 0, 999999999, time.UTC)
	j.CreatedAt = stamp
	j.UpdatedAt = stamp
	j.ReviewedAt = &stamp
	b, e := json.Marshal(j)
	return len(b), e
}

func WithinSizeLimit(j Job) bool {
	n, e := ProjectedBytes(j)
	return e == nil && n+1024 <= MaxDraftBytes-4096
}

type Source struct {
	PDF                            PDFEvidence // Nil for legacy requests supplying page text.
	SourceDocument, SourceRevision string
	Pages                          []Page
}

// ValidateSource is shared by manual candidates and model parsing; it does not retain text.
func ValidateSource(input Source) (Source, string, error) {
	j := Source{PDF: input.PDF, SourceDocument: strings.TrimSpace(input.SourceDocument), SourceRevision: strings.TrimSpace(input.SourceRevision), Pages: append([]Page{}, input.Pages...)}
	if j.SourceDocument == "" || !utf8.ValidString(j.SourceDocument) || utf8.RuneCountInString(j.SourceDocument) > 200 || !utf8.ValidString(j.SourceRevision) || utf8.RuneCountInString(j.SourceRevision) > 80 || len(input.Pages) < 1 || len(input.Pages) > 50 {
		return Source{}, "", ErrInput
	}
	pages := map[int]string{}
	total := 0
	for _, p := range input.Pages {
		total += len(p.Text)
		if p.Number < 1 || p.Number > 1000000 || !utf8.ValidString(p.Text) || len(p.Text) > 32*1024 || total > 128*1024 {
			return Source{}, "", ErrInput
		}
		if _, ok := pages[p.Number]; ok {
			return Source{}, "", ErrInput
		}
		pages[p.Number] = p.Text
	}
	canonicalPages := append([]Page{}, input.Pages...)
	slices.SortFunc(canonicalPages, func(a, b Page) int { return a.Number - b.Number })
	type sourcePageV1 struct {
		Number int
		Text   string
	}
	pagesV1 := make([]sourcePageV1, 0, len(canonicalPages))
	for _, p := range canonicalPages {
		pagesV1 = append(pagesV1, sourcePageV1{Number: p.Number, Text: p.Text})
	}
	digest := Hash(struct {
		Version int
		Pages   []sourcePageV1
	}{1, pagesV1})
	return j, digest, nil
}

func Prepare(input Creation) (Job, string, error) {
	source, digest, e := ValidateSource(Source{SourceDocument: input.SourceDocument, SourceRevision: input.SourceRevision, Pages: input.Pages})
	if e != nil || len(input.Rows) < 1 || len(input.Rows) > 100 {
		return Job{}, "", ErrInput
	}
	j := Job{SourceDocument: source.SourceDocument, SourceRevision: source.SourceRevision, SourceHash: digest, Status: "draft", Version: 1, Rows: []Row{}}
	pages := map[int]string{}
	for _, p := range source.Pages {
		pages[p.Number] = p.Text
	}
	seen := map[string]bool{}
	for _, candidate := range input.Rows {
		r := Row{Fields: candidate.Fields, SourceItemNo: candidate.SourceItemNo, Evidence: append([]Evidence{}, candidate.Evidence...)}
		if len(r.Evidence) < 1 || len(r.Evidence) > 4 {
			return Job{}, "", ErrInput
		}
		for _, e := range r.Evidence {
			text, ok := pages[e.Page]
			if !ok || strings.TrimSpace(e.Quote) == "" || !utf8.ValidString(e.Quote) || utf8.RuneCountInString(e.Quote) > 500 || !strings.Contains(text, e.Quote) {
				return Job{}, "", ErrInput
			}
		}
		r, e := normalize(j, r)
		if e != nil || r.SourceItemNo == "" || seen[r.SourceItemNo] {
			return Job{}, "", ErrInput
		}
		seen[r.SourceItemNo] = true
		j.Rows = append(j.Rows, r)
	}
	payload := creationHash(j)
	if !WithinSizeLimit(j) {
		return Job{}, "", ErrInput
	}
	return j, payload, nil
}

func ApplyReview(j Job, actor string, input Review, now time.Time) (Job, error) {
	if actor == "" || actor != j.CreatorID {
		return Job{}, ErrForbidden
	}
	if (j.Status != "draft" && j.Status != "reviewed") || j.Version != input.ExpectedVersion || j.Version >= repositoryDomainEntity.MaxResourceNumber {
		return Job{}, ErrConflict
	}
	if len(input.Rows) != len(j.Rows) {
		return Job{}, ErrInput
	}
	edits := map[string]Fields{}
	for _, r := range input.Rows {
		if _, ok := edits[r.ID]; ok {
			return Job{}, ErrInput
		}
		edits[r.ID] = r.Fields
	}
	j.Rows = append([]Row{}, j.Rows...)
	for i, r := range j.Rows {
		f, ok := edits[r.ID]
		if !ok {
			return Job{}, ErrInput
		}
		r.Fields = f
		var e error
		r, e = normalize(j, r)
		if e != nil {
			return Job{}, e
		}
		j.Rows[i] = r
	}
	j.Status = "reviewed"
	j.Version++
	j.UpdatedAt = now
	j.ReviewedAt = &now
	if !WithinSizeLimit(j) {
		return Job{}, ErrInput
	}
	return j, nil
}

func Begin(j Job, actor string, expected int64, now time.Time) (Job, error) {
	if actor == "" || actor != j.CreatorID {
		return Job{}, ErrForbidden
	}
	if j.Status == "confirming" || j.Status == "completed" {
		if expected != j.ConfirmedFromReviewVersion {
			return Job{}, ErrConflict
		}
		return j, nil
	}
	if j.Status != "reviewed" || j.Version != expected || j.Version > repositoryDomainEntity.MaxResourceNumber-1-int64(len(j.Rows)) {
		return Job{}, ErrConflict
	}
	for _, r := range j.Rows {
		if _, e := normalize(j, r); e != nil {
			return Job{}, e
		}
	}
	j.ConfirmedFromReviewVersion = expected
	j.Status = "confirming"
	j.Version++
	j.UpdatedAt = now
	return j, nil
}

func Record(j Job, actor, rowID, objectID string, reviewVersion int64, now time.Time) (Job, error) {
	if actor == "" || actor != j.CreatorID {
		return Job{}, ErrForbidden
	}
	if (j.Status != "confirming" && j.Status != "completed") || j.ConfirmedFromReviewVersion != reviewVersion || !repositoryDomainEntity.ValidID(objectID) {
		return Job{}, ErrConflict
	}
	j.Rows = append([]Row{}, j.Rows...)
	found := false
	all := true
	for i, r := range j.Rows {
		if r.ID == rowID {
			found = true
			if r.ResultObjectID != "" {
				if r.ResultObjectID != objectID {
					return Job{}, ErrConflict
				}
				return j, nil
			}
			j.Rows[i].ResultObjectID = objectID
		}
		all = all && j.Rows[i].ResultObjectID != ""
	}
	if !found {
		return Job{}, ErrInput
	}
	if j.Version >= repositoryDomainEntity.MaxResourceNumber {
		return Job{}, ErrConflict
	}
	j.Version++
	j.UpdatedAt = now
	if all {
		j.Status = "completed"
	}
	return j, nil
}

func ValidateStored(j Job) error {
	if strings.TrimSpace(j.SourceDocument) != j.SourceDocument || j.SourceDocument == "" || !utf8.ValidString(j.SourceDocument) || utf8.RuneCountInString(j.SourceDocument) > 200 || strings.TrimSpace(j.SourceRevision) != j.SourceRevision || !utf8.ValidString(j.SourceRevision) || utf8.RuneCountInString(j.SourceRevision) > 80 || !hashPattern.MatchString(j.SourceHash) {
		return ErrUnavailable
	}
	if j.Status != "draft" && (j.ReviewedAt == nil || j.ReviewedAt.Before(j.CreatedAt) || j.ReviewedAt.After(j.UpdatedAt)) {
		return ErrUnavailable
	}
	if !repositoryDomainEntity.ValidID(j.ID) || !repositoryDomainEntity.ValidID(j.RepositoryID) || j.CreatorID == "" || len(j.Rows) < 1 || len(j.Rows) > 100 || j.Version < 1 || j.Version > repositoryDomainEntity.MaxResourceNumber || j.CreatedAt.IsZero() || j.UpdatedAt.Before(j.CreatedAt) || !WithinSizeLimit(j) {
		return ErrUnavailable
	}
	switch j.Status {
	case "draft", "reviewed":
		if (j.Status == "reviewed" && j.Version < 2) || (j.Status == "draft" && j.ReviewedAt != nil) || j.ConfirmedFromReviewVersion != 0 {
			return ErrUnavailable
		}
	case "confirming", "completed":
		if j.ConfirmedFromReviewVersion < 2 || j.ConfirmedFromReviewVersion >= j.Version {
			return ErrUnavailable
		}
	default:
		return ErrUnavailable
	}
	ids := map[string]bool{}
	items := map[string]bool{}
	all := true
	for _, r := range j.Rows {
		if !repositoryDomainEntity.ValidID(r.ID) || ids[r.ID] || items[r.SourceItemNo] || r.SourceItemNo == "" || len(r.Evidence) < 1 || len(r.Evidence) > 4 {
			return ErrUnavailable
		}
		ids[r.ID] = true
		items[r.SourceItemNo] = true
		n, e := normalize(j, r)
		if e != nil || n.Fields != r.Fields || n.SourceItemNo != r.SourceItemNo {
			return ErrUnavailable
		}
		for _, e := range r.Evidence {
			if e.Page < 1 || e.Page > 1000000 || !utf8.ValidString(e.Quote) || strings.TrimSpace(e.Quote) == "" || utf8.RuneCountInString(e.Quote) > 500 {
				return ErrUnavailable
			}
		}
		if r.ResultObjectID != "" && !repositoryDomainEntity.ValidID(r.ResultObjectID) {
			return ErrUnavailable
		}
		if r.ResultObjectID != "" && (j.Status == "draft" || j.Status == "reviewed") {
			return ErrUnavailable
		}
		all = all && r.ResultObjectID != ""
	}
	if (j.Status == "completed" && !all) || (j.Status == "confirming" && all) {
		return ErrUnavailable
	}
	return nil
}

// creationHash freezes the v1 intent independently of job lifecycle/schema fields.
func creationHash(j Job) string {
	type proof struct {
		Page  int
		Quote string
	}
	type line struct {
		Name, Description, Type, Quantity, Unit, MaterialCode, SourceItemNo string
		Evidence                                                            []proof
	}
	rows := make([]line, 0, len(j.Rows))
	for _, r := range j.Rows {
		p := make([]proof, 0, len(r.Evidence))
		for _, e := range r.Evidence {
			p = append(p, proof(e))
		}
		rows = append(rows, line{r.Name, r.Description, r.Type, r.Quantity, r.Unit, r.MaterialCode, r.SourceItemNo, p})
	}
	return Hash(struct {
		SchemaVersion                              int
		SourceDocument, SourceRevision, SourceHash string
		Rows                                       []line
	}{1, j.SourceDocument, j.SourceRevision, j.SourceHash, rows})
}
