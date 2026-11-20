package entity

import (
	"errors"
	"strings"
	"time"

	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

var ErrInput = errors.New("invalid subobject")

var ErrNotFound = errors.New("subobject or parent not found")

var ErrForbidden = errors.New("parent owner required")

var ErrConflict = errors.New("subobject version conflict")

var ErrCreationConflict = errors.New("subobject intent conflict")

var ErrProgressRemoval = errors.New("completed subobject cannot become non-work")

var ErrUnavailable = errors.New("subobject unavailable")

type Content struct{ Name, Type, Quantity, Unit, CompletedQuantity string }

type Subobject struct {
	Content
	ID, RepositoryID, ParentObjectID, CreatorID, Status string
	Version                                             int64
	CreatedAt, UpdatedAt                                time.Time
}

type Update struct {
	Content
	CompletedQuantity *string
	ExpectedVersion   int64
}

// Normalize reuses exact decimal and text rules without inheriting Object source fields.
func Normalize(c Content) (Content, error) {
	c.Type = strings.TrimSpace(c.Type)
	if c.Type != objectDomainEntity.TypeWork && c.Type != objectDomainEntity.TypeNonWork {
		return Content{}, ErrInput
	}
	n, e := objectDomainEntity.Normalize(objectDomainEntity.Content{Name: c.Name, Type: c.Type, Quantity: c.Quantity, Unit: c.Unit, CompletedQuantity: c.CompletedQuantity})
	if e != nil {
		return Content{}, ErrInput
	}
	return Content{Name: n.Name, Type: n.Type, Quantity: n.Quantity, Unit: n.Unit, CompletedQuantity: n.CompletedQuantity}, nil
}

func Status(c Content) (string, error) {
	if c.Type == objectDomainEntity.TypeNonWork && c.CompletedQuantity == "" {
		return objectDomainEntity.StatusNotStarted, nil
	}
	status, e := objectDomainEntity.WorkStatus(c.Quantity, c.CompletedQuantity)
	if e != nil {
		return "", ErrInput
	}
	return status, nil
}

func (v Subobject) RequiresProgressReset() bool {
	return v.Type == objectDomainEntity.TypeWork && v.CompletedQuantity != "0"
}

func ApplyUpdate(current Subobject, actor string, input Update) (Subobject, error) {
	if actor == "" || actor != current.CreatorID {
		return Subobject{}, ErrForbidden
	}
	c := input.Content
	c.Type = strings.TrimSpace(c.Type)
	if input.CompletedQuantity != nil {
		c.CompletedQuantity = *input.CompletedQuantity
	} else if c.Type == objectDomainEntity.TypeWork && current.Type == objectDomainEntity.TypeWork {
		c.CompletedQuantity = current.CompletedQuantity
	} else {
		c.CompletedQuantity = ""
	}
	if c.Type == objectDomainEntity.TypeNonWork && current.RequiresProgressReset() {
		return Subobject{}, ErrProgressRemoval
	}
	if c.Type == objectDomainEntity.TypeWork && c.CompletedQuantity == "" {
		return Subobject{}, ErrInput
	}
	normalized, e := Normalize(c)
	if e != nil {
		return Subobject{}, e
	}
	status, e := Status(normalized)
	if e != nil || input.ExpectedVersion < 1 {
		return Subobject{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Subobject{}, ErrConflict
	}
	if current.Content == normalized && current.Status == status {
		return current, nil
	}
	if current.Version >= repositoryDomainEntity.MaxResourceNumber {
		return Subobject{}, ErrConflict
	}
	current.Content = normalized
	current.Status = status
	current.Version++
	return current, nil
}
