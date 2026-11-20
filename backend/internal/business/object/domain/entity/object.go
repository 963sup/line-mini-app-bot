package entity

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var ErrProgressRemoval = errors.New("completed work cannot become non-work")

var ErrInput = errors.New("invalid object input")

var ErrNotFound = errors.New("object not found")

var ErrForbidden = errors.New("only the creator may update this object")

var ErrConflict = errors.New("object version conflict")

var ErrCreationConflict = errors.New("object creation key conflict")

var ErrUnavailable = errors.New("object storage unavailable")

var decimal = regexp.MustCompile(`^(0|[1-9][0-9]{0,17})(\.[0-9]{1,6})?$`)

type Content struct {
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
	Content
	ID, RepositoryID, CreatorID, Status string
	Version                             int64
	CreatedAt, UpdatedAt                time.Time
}

type SourceUpdate struct{ SourceDocument, SourceRevision, SourceItemNo, SourcePages, MaterialCode *string }

type Update struct {
	Content
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

// WorkStatus uses exact rational comparison after decimal validation.
func WorkStatus(total, done string) (string, error) {
	if !decimal.MatchString(total) || !decimal.MatchString(done) {
		return "", ErrInput
	}
	quantity, ok := new(big.Rat).SetString(total)
	if !ok || quantity.Sign() <= 0 {
		return "", ErrInput
	}
	completed, ok := new(big.Rat).SetString(done)
	if !ok || completed.Sign() < 0 || completed.Cmp(quantity) > 0 {
		return "", ErrInput
	}
	switch {
	case completed.Sign() == 0:
		return StatusNotStarted, nil
	case completed.Cmp(quantity) == 0:
		return StatusCompleted, nil
	default:
		return StatusInProgress, nil
	}
}

func canonicalDecimal(s string) string {
	if strings.Contains(s, ".") {
		return strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

func Normalize(c Content) (Content, error) {
	c.SpaceID = strings.TrimSpace(c.SpaceID)
	if c.SpaceID != "" && !repository.ValidID(c.SpaceID) {
		return Content{}, ErrInput
	}
	fields := []struct {
		v        *string
		limit    int
		required bool
	}{{&c.Name, 120, true}, {&c.Description, 8000, false}, {&c.Type, 80, true}, {&c.Unit, 32, true}, {&c.SourceDocument, 200, false}, {&c.SourceRevision, 80, false}, {&c.SourceItemNo, 80, false}, {&c.SourcePages, 80, false}, {&c.MaterialCode, 120, false}}
	for _, f := range fields {
		*f.v = strings.TrimSpace(*f.v)
		if !utf8.ValidString(*f.v) || utf8.RuneCountInString(*f.v) > f.limit || (f.required && *f.v == "") {
			return Content{}, ErrInput
		}
	}
	c.Quantity = strings.TrimSpace(c.Quantity)
	if !decimal.MatchString(c.Quantity) {
		return Content{}, ErrInput
	}
	if strings.Contains(c.Quantity, ".") {
		c.Quantity = strings.TrimRight(strings.TrimRight(c.Quantity, "0"), ".")
	}
	if c.Quantity == "0" {
		return Content{}, ErrInput
	}
	if c.CompletedQuantity != "" {
		c.CompletedQuantity = strings.TrimSpace(c.CompletedQuantity)
		if c.Type != TypeWork {
			return Content{}, ErrInput
		}
		if _, e := WorkStatus(c.Quantity, c.CompletedQuantity); e != nil {
			return Content{}, e
		}
		c.CompletedQuantity = canonicalDecimal(c.CompletedQuantity)
	}
	return c, nil
}

func ApplyUpdate(current Object, actor string, input Update) (Object, error) {
	if actor == "" || actor != current.CreatorID {
		return Object{}, ErrForbidden
	}
	content := input.Content
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
	if content.Type == TypeWork && current.Type != TypeWork && (input.CompletedQuantity == nil || strings.TrimSpace(*input.CompletedQuantity) == "") {
		return Object{}, ErrInput
	}
	if content.Type == TypeNonWork && input.CompletedQuantity != nil && strings.TrimSpace(*input.CompletedQuantity) != "" {
		return Object{}, ErrInput
	}
	if current.Type == TypeWork && current.CompletedQuantity != "" && content.Type != TypeWork {
		done, _ := new(big.Rat).SetString(current.CompletedQuantity)
		if done == nil || done.Sign() > 0 {
			return Object{}, ErrProgressRemoval
		}
		content.CompletedQuantity = ""
	}
	if content.Type == TypeWork && input.CompletedQuantity != nil && strings.TrimSpace(*input.CompletedQuantity) == "" {
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
	if current.Version >= repository.MaxResourceNumber {
		return Object{}, ErrConflict
	}
	current.Content = c
	current.Status = status
	current.Version++
	return current, nil
}
