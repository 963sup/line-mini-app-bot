package milestones

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

var ErrInput = errors.New("invalid milestones input")

var ErrNotFound = errors.New("milestones not found")

var ErrForbidden = errors.New("only creator may edit")

var ErrConflict = errors.New("version conflict")

var ErrCreationConflict = errors.New("creation key conflict")

var ErrUnavailable = errors.New("milestones unavailable")

type Milestone struct {
	ID, RepositoryID, Title, Body, CreatorID string
	Version                                  int64
	CreatedAt, UpdatedAt                     time.Time
	IssueCount, CompletedIssueCount          int64
	State, DueDate                           string
}

type Content struct {
	Title, Body string
	DueDate     string
}

type Update struct {
	Content
	ExpectedVersion int64
	State           string
}

func Normalize(c Content) (Content, error) {
	c.Title = strings.TrimSpace(c.Title)
	c.Body = strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	if c.DueDate != "" {
		if _, e := time.Parse("2006-01-02", c.DueDate); e != nil {
			return Content{}, ErrInput
		}
	}
	return c, nil
}

func CheckUpdate(current Milestone, actor string, version int64) error {
	if actor == "" || actor != current.CreatorID {
		return ErrForbidden
	}
	if version < 1 {
		return ErrInput
	}
	if current.Version != version {
		return ErrConflict
	}
	return nil
}

func ApplyUpdate(current Milestone, actor string, input Update) (Milestone, error) {
	if e := CheckUpdate(current, actor, input.ExpectedVersion); e != nil {
		return Milestone{}, e
	}
	c, e := Normalize(input.Content)
	if e != nil {
		return Milestone{}, e
	}
	if input.State != "open" && input.State != "completed" {
		return Milestone{}, ErrInput
	}
	if c.Title == current.Title && c.Body == current.Body && c.DueDate == current.DueDate && input.State == current.State {
		return current, nil
	}
	if current.Version >= repositoryDomainEntity.MaxResourceNumber {
		return Milestone{}, ErrConflict
	}
	current.Title = c.Title
	current.Body = c.Body
	current.DueDate = c.DueDate
	current.State = input.State
	current.Version++
	return current, nil
}
