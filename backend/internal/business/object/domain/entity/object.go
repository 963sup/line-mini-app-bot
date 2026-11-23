package entity

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/policy"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/valueobject"
)

var ErrProgressRemoval = errors.New("completed work cannot become non-work")

var ErrInput = errors.New("invalid object input")

var ErrNotFound = errors.New("object not found")

var ErrForbidden = errors.New("only the creator may update this object")

var ErrConflict = errors.New("object version conflict")

var ErrCreationConflict = errors.New("object creation key conflict")

var ErrUnavailable = errors.New("object storage unavailable")

const maxObjectVersion int64 = 1<<53 - 1 // Largest exact integer in the public JSON contract.

var resourceID = regexp.MustCompile(`^[0-9a-f]{32}$`)

type Content struct {
	ParentID          string
	SpaceID           string
	Name              string
	Description       string
	Type              string
	Quantity          string
	Unit              string
	SourceDocument    string
	SourceRevision    string
	SourceItemNo      string
	SourcePages       string
	MaterialCode      string
	CompletedQuantity string
}

type Object struct {
	TemplateID, TemplateNodeID string
	TemplateVersion            int64
	Content
	ID, RepositoryID, CreatorID, Status string
	Version                             int64
	CreatedAt, UpdatedAt                time.Time
}

// New validates the prepared content for a manual object or a template node.
// Creation defaults are applied before persistence so legacy inputs retain absent completion.
func New(id, repositoryID, creatorID string, content Content, at time.Time) (Object, error) {
	content, err := Normalize(content)
	if err != nil {
		return Object{}, err
	}
	return Object{
		Content: content, ID: id, RepositoryID: repositoryID, CreatorID: creatorID,
		Status: content.InitialStatus(), Version: 1, CreatedAt: at, UpdatedAt: at,
	}, nil
}

func (v Object) CanEdit(actor string) bool { return policy.CanEdit(actor, v.CreatorID) }

type SourceUpdate struct{ SourceDocument, SourceRevision, SourceItemNo, SourcePages, MaterialCode *string }

type Update struct {
	Content
	ActorDisplayName  string // Authenticated identity metadata; never populated from the update body.
	Status            string
	ExpectedVersion   int64
	CompletedQuantity *string
	SpaceID           *string
	Sources           SourceUpdate
}

const TypeWork = "work"

const TypeNonWork = "non_work"

const StatusNotStarted = "not_started"

const StatusInProgress = "in_progress"

const StatusCompleted = "completed"

func ValidStatus(s string) bool {
	switch s {
	case StatusNotStarted, StatusInProgress, StatusCompleted:
		return true
	default:
		return false
	}
}

// WorkStatus compares exact decimal quantities.
func WorkStatus(total, done string) (string, error) {
	quantity, ok := valueobject.ParseQuantity(total)
	completed, valid := valueobject.ParseQuantity(done)
	if !ok || !valid || quantity.IsZero() || completed.Compare(quantity) > 0 {
		return "", ErrInput
	}
	switch {
	case completed.IsZero():
		return StatusNotStarted, nil
	case completed.Compare(quantity) == 0:
		return StatusCompleted, nil
	default:
		return StatusInProgress, nil
	}
}

// NormalizeCreation supplies the initial completed quantity for new tracked work.
func NormalizeCreation(c Content) (Content, error) {
	if strings.TrimSpace(c.Type) == TypeWork && strings.TrimSpace(c.Quantity) != "" && c.CompletedQuantity == "" {
		c.CompletedQuantity = "0"
	}
	return Normalize(c)
}

// InitialStatus derives the initial state of normalized creation content.
func (c Content) InitialStatus() string {
	if c.Type == TypeWork && c.CompletedQuantity != "" {
		status, _ := WorkStatus(c.Quantity, c.CompletedQuantity)
		return status
	}
	return StatusNotStarted
}

func Normalize(c Content) (Content, error) {
	if c.ParentID != "" && !resourceID.MatchString(c.ParentID) {
		return Content{}, ErrInput
	}
	c.SpaceID = strings.TrimSpace(c.SpaceID)
	if c.SpaceID != "" && !resourceID.MatchString(c.SpaceID) {
		return Content{}, ErrInput
	}
	fields := []struct {
		v        *string
		limit    int
		required bool
	}{{&c.Name, 120, true}, {&c.Description, 8000, false}, {&c.Type, 80, false}, {&c.Unit, 32, false}, {&c.SourceDocument, 200, false}, {&c.SourceRevision, 80, false}, {&c.SourceItemNo, 80, false}, {&c.SourcePages, 80, false}, {&c.MaterialCode, 120, false}}
	for _, f := range fields {
		*f.v = strings.TrimSpace(*f.v)
		if !utf8.ValidString(*f.v) || utf8.RuneCountInString(*f.v) > f.limit || (f.required && *f.v == "") {
			return Content{}, ErrInput
		}
	}
	c.Quantity = strings.TrimSpace(c.Quantity)
	if c.Quantity != "" {
		q, ok := valueobject.ParseQuantity(c.Quantity)
		if !ok || q.IsZero() {
			return Content{}, ErrInput
		}
		c.Quantity = q.String()
	}
	if c.CompletedQuantity != "" {
		c.CompletedQuantity = strings.TrimSpace(c.CompletedQuantity)
		if c.Type != TypeWork || c.Unit == "" {
			return Content{}, ErrInput
		}
		if _, e := WorkStatus(c.Quantity, c.CompletedQuantity); e != nil {
			return Content{}, e
		}
		done, _ := valueobject.ParseQuantity(c.CompletedQuantity)
		c.CompletedQuantity = done.String()
	}
	return c, nil
}

// WorkConversionDefault preserves legacy progress when switching to measured work.
func (v Object) WorkConversionDefault() string {
	if v.Type != TypeWork && v.Type != TypeNonWork {
		switch v.Status {
		case StatusNotStarted:
			return "zero"
		case StatusCompleted:
			return "quantity"
		}
	}
	return ""
}

func (v Object) RequiresProgressReset() bool {
	if v.Type != TypeWork || v.CompletedQuantity == "" {
		return false
	}
	done, _ := new(big.Rat).SetString(v.CompletedQuantity)
	return done == nil || done.Sign() > 0
}

func ApplyUpdate(current Object, actor string, input Update) (Object, error) {
	if !current.CanEdit(actor) {
		return Object{}, ErrForbidden
	}
	content := input.Content
	content.ParentID = current.ParentID
	content.SpaceID = current.SpaceID
	if input.SpaceID != nil {
		content.SpaceID = *input.SpaceID
	}
	content.Type = strings.TrimSpace(content.Type)
	for _, f := range []struct {
		value  *string
		update *string
		prior  string
	}{{&content.SourceDocument, input.Sources.SourceDocument, current.SourceDocument}, {&content.SourceRevision, input.Sources.SourceRevision, current.SourceRevision}, {&content.SourceItemNo, input.Sources.SourceItemNo, current.SourceItemNo}, {&content.SourcePages, input.Sources.SourcePages, current.SourcePages}, {&content.MaterialCode, input.Sources.MaterialCode, current.MaterialCode}} {
		if f.update == nil {
			*f.value = f.prior
		} else {
			*f.value = *f.update
		}
	}
	// Progress is absolute, and omission preserves already managed work.
	if input.CompletedQuantity != nil {
		content.CompletedQuantity = *input.CompletedQuantity
	} else if content.Type == TypeWork && current.Type == TypeWork {
		content.CompletedQuantity = current.CompletedQuantity
	} else {
		content.CompletedQuantity = ""
	}
	if content.Type == TypeWork && content.Quantity != "" && current.Type != TypeWork && (input.CompletedQuantity == nil || strings.TrimSpace(*input.CompletedQuantity) == "") {
		return Object{}, ErrInput
	}
	if content.Type == TypeNonWork && input.CompletedQuantity != nil && strings.TrimSpace(*input.CompletedQuantity) != "" {
		return Object{}, ErrInput
	}
	if current.Type == TypeWork && current.CompletedQuantity != "" && content.Type != TypeWork {
		if current.RequiresProgressReset() {
			return Object{}, ErrProgressRemoval
		}
		content.CompletedQuantity = ""
	}
	if content.Type == TypeWork && input.CompletedQuantity != nil && strings.TrimSpace(*input.CompletedQuantity) == "" && (content.Quantity != "" || current.CompletedQuantity != "") {
		return Object{}, ErrInput
	}
	c, err := Normalize(content)
	status := input.Status
	switch {
	case c.Type == TypeWork && c.CompletedQuantity != "":
		derived, e := WorkStatus(c.Quantity, c.CompletedQuantity)
		if e != nil {
			return Object{}, e
		}
		if status != "" && status != derived {
			return Object{}, ErrInput
		}
		status = derived
	case status == "":
		status = current.Status
	}
	if err != nil || !ValidStatus(status) || input.ExpectedVersion < 1 {
		return Object{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Object{}, ErrConflict
	}
	if current.Content == c && current.Status == status {
		return current, nil
	}
	if current.Version >= maxObjectVersion {
		return Object{}, ErrConflict
	}
	current.Content = c
	current.Status = status
	current.Version++
	return current, nil
}
