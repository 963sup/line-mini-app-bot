// Package objectimport validates untrusted candidate rows and owns reviewed import drafts.
package objectimport

import (
	"context"
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

	"github.com/963sup/line-mini-app-bot/internal/business/object"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrInput            = errors.New("invalid object import")
	ErrNotFound         = errors.New("object import not found")
	ErrForbidden        = errors.New("object import owner required")
	ErrConflict         = errors.New("object import version or state conflict")
	ErrCreationConflict = errors.New("object import creation key conflict")
	ErrUnavailable      = errors.New("object import unavailable")
	hashPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	keyPattern          = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

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
type Page struct {
	Number int
	Text   string
}
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

func newID() (string, error) {
	var b [16]byte
	_, e := rand.Read(b[:])
	return hex.EncodeToString(b[:]), e
}

func objectContent(j Job, r Row) object.Content {
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
	return object.Content{Name: r.Name, Description: r.Description, Type: r.Type, Quantity: r.Quantity, Unit: r.Unit, MaterialCode: r.MaterialCode, SourceDocument: j.SourceDocument, SourceRevision: j.SourceRevision, SourceItemNo: r.SourceItemNo, SourcePages: strings.Join(labels, ",")}
}

func normalize(j Job, r Row) (Row, error) {
	c, e := object.Normalize(objectContent(j, r))
	if e != nil {
		return Row{}, ErrInput
	}
	r.Fields = Fields{c.Name, c.Description, c.Type, c.Quantity, c.Unit, c.MaterialCode}
	r.SourceItemNo = c.SourceItemNo
	return r, nil
}

// Budget a fixed final snapshot, so later review/result metadata consumes reserved
// capacity instead of reapplying a moving threshold to the same accepted draft.
func projectedBytes(j Job) (int, error) {
	j.Rows = append([]Row{}, j.Rows...)
	for i := range j.Rows {
		j.Rows[i].ID = strings.Repeat("f", 32)
		j.Rows[i].ResultObjectID = strings.Repeat("f", 32)
	}
	j.ID = strings.Repeat("f", 32)
	j.RepositoryID = strings.Repeat("f", 32)
	j.Status = "completed"
	j.Version = repository.MaxResourceNumber
	j.ConfirmedFromReviewVersion = repository.MaxResourceNumber
	stamp := time.Date(2026, 1, 1, 0, 0, 0, 999999999, time.UTC)
	j.CreatedAt = stamp
	j.UpdatedAt = stamp
	j.ReviewedAt = &stamp
	b, e := json.Marshal(j)
	return len(b), e
}

func bounded(j Job) bool { n, e := projectedBytes(j); return e == nil && n+1024 <= MaxDraftBytes-4096 }

type Source struct {
	SourceDocument, SourceRevision string
	Pages                          []Page
}

// ValidateSource is shared by manual candidates and model parsing; it does not retain text.
func ValidateSource(input Source) (Source, string, error) {
	j := Source{SourceDocument: strings.TrimSpace(input.SourceDocument), SourceRevision: strings.TrimSpace(input.SourceRevision), Pages: append([]Page{}, input.Pages...)}
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
		pagesV1 = append(pagesV1, sourcePageV1(p))
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
	if !bounded(j) {
		return Job{}, "", ErrInput
	}
	return j, payload, nil
}

func ApplyReview(j Job, actor string, input Review, now time.Time) (Job, error) {
	if actor == "" || actor != j.CreatorID {
		return Job{}, ErrForbidden
	}
	if (j.Status != "draft" && j.Status != "reviewed") || j.Version != input.ExpectedVersion || j.Version >= repository.MaxResourceNumber {
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
	if !bounded(j) {
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
	if j.Status != "reviewed" || j.Version != expected || j.Version > repository.MaxResourceNumber-1-int64(len(j.Rows)) {
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
	if (j.Status != "confirming" && j.Status != "completed") || j.ConfirmedFromReviewVersion != reviewVersion || !repository.ValidID(objectID) {
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
	if j.Version >= repository.MaxResourceNumber {
		return Job{}, ErrConflict
	}
	j.Version++
	j.UpdatedAt = now
	if all {
		j.Status = "completed"
	}
	return j, nil
}

type Repositories interface {
	Get(context.Context, string, string) (repository.Repository, error)
}
type Objects interface {
	Create(context.Context, string, string, string, object.Content) (object.Object, error)
}
type Store interface {
	Create(context.Context, repository.Repository, string, string, string, Job) (Job, error)
	Get(context.Context, repository.Repository, string, string) (Job, error)
	Review(context.Context, repository.Repository, string, string, Review) (Job, error)
	BeginConfirm(context.Context, repository.Repository, string, string, int64) (Job, error)
	RecordResult(context.Context, repository.Repository, string, string, string, string, int64) (Job, error)
}
type Service struct {
	Repositories Repositories
	Objects      Objects
	Store        Store
}

func (s *Service) scope(ctx context.Context, actor, rid string) (repository.Repository, error) {
	if s == nil || s.Store == nil || s.Repositories == nil {
		return repository.Repository{}, ErrUnavailable
	}
	if actor == "" || !repository.ValidID(rid) {
		return repository.Repository{}, ErrNotFound
	}
	r, e := s.Repositories.Get(ctx, actor, rid)
	if errors.Is(e, repository.ErrNotFound) {
		return repository.Repository{}, ErrNotFound
	}
	if e != nil {
		return repository.Repository{}, ErrUnavailable
	}
	if r.ID != rid || r.Status != "ACTIVE" || !repository.ValidGroup(r.GroupID) {
		return repository.Repository{}, ErrNotFound
	}
	return r, nil
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, input Creation) (Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Job{}, e
	}
	if !keyPattern.MatchString(key) {
		return Job{}, ErrInput
	}
	j, payload, e := Prepare(input)
	if e != nil {
		return Job{}, e
	}
	j.ID, e = newID()
	if e != nil {
		return Job{}, ErrUnavailable
	}
	j.RepositoryID = rid
	j.CreatorID = actor
	j.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	j.UpdatedAt = j.CreatedAt
	for i := range j.Rows {
		j.Rows[i].ID, e = newID()
		if e != nil {
			return Job{}, ErrUnavailable
		}
	}
	if !bounded(j) {
		return Job{}, ErrInput
	}
	return s.Store.Create(ctx, r, actor, key, payload, j)
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Job{}, e
	}
	if !repository.ValidID(id) {
		return Job{}, ErrNotFound
	}
	return s.Store.Get(ctx, r, actor, id)
}

func (s *Service) Review(ctx context.Context, actor, rid, id string, input Review) (Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Job{}, e
	}
	if !repository.ValidID(id) {
		return Job{}, ErrNotFound
	}
	return s.Store.Review(ctx, r, actor, id, input)
}

func (s *Service) Confirm(ctx context.Context, actor, rid, id string, expected int64) (Job, error) {
	r, e := s.scope(ctx, actor, rid)
	if e != nil {
		return Job{}, e
	}
	if !repository.ValidID(id) {
		return Job{}, ErrNotFound
	}
	if s.Objects == nil {
		return Job{}, ErrUnavailable
	}
	j, e := s.Store.BeginConfirm(ctx, r, actor, id, expected)
	if e != nil {
		return Job{}, e
	}
	for _, row := range j.Rows {
		if row.ResultObjectID != "" {
			continue
		}
		key := Hash([]string{"object-import-v1", j.ID, row.ID})
		created, e := s.Objects.Create(ctx, actor, rid, key, objectContent(j, row))
		if e != nil {
			return j, ErrUnavailable
		}
		if !repository.ValidID(created.ID) {
			return j, ErrUnavailable
		}
		j, e = s.Store.RecordResult(ctx, r, actor, id, row.ID, created.ID, expected)
		if e != nil {
			return Job{}, e
		}
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
	if !repository.ValidID(j.ID) || !repository.ValidID(j.RepositoryID) || j.CreatorID == "" || len(j.Rows) < 1 || len(j.Rows) > 100 || j.Version < 1 || j.Version > repository.MaxResourceNumber || j.CreatedAt.IsZero() || j.UpdatedAt.Before(j.CreatedAt) || !bounded(j) {
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
		if !repository.ValidID(r.ID) || ids[r.ID] || items[r.SourceItemNo] || r.SourceItemNo == "" || len(r.Evidence) < 1 || len(r.Evidence) > 4 {
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
		if r.ResultObjectID != "" && !repository.ValidID(r.ResultObjectID) {
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
