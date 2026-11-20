package entity

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInput = errors.New("invalid project input")

var ErrNotFound = errors.New("project not found")

var ErrForbidden = errors.New("only creator may edit")

var ErrConflict = errors.New("version conflict")

var ErrCreationConflict = errors.New("creation key conflict")

var ErrUnavailable = errors.New("project unavailable")

var ErrSourceUnavailable = errors.New("project source unavailable")

type Project struct {
	ID, DefaultRepositoryID, OwnerID, Title, Body, CreatorID string
	LinkedRepositoryIDs                                      []string
	SourceAvailable                                          bool
	Version                                                  int64
	CreatedAt, UpdatedAt                                     time.Time
	IssueCount, CompletedIssueCount                          int64
}

type Content struct{ Title, Body, SourceRepositoryID string }

type Update struct {
	Content
	ExpectedVersion                      int64
	LinkRepositoryID, UnlinkRepositoryID string
	ItemID, ItemStatus                   string
}

func Normalize(c Content) (Content, error) {
	c.Title = strings.TrimSpace(c.Title)
	c.Body = strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	return c, nil
}

func CheckUpdate(current Project, actor string, version int64) error {
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

func ApplyUpdate(current Project, actor string, input Update) (Project, error) {
	if e := CheckUpdate(current, actor, input.ExpectedVersion); e != nil {
		return Project{}, e
	}
	c, e := Normalize(input.Content)
	if e != nil {
		return Project{}, e
	}
	if c.Title == current.Title && c.Body == current.Body {
		return current, nil
	}
	if current.Version >= MaxVersion {
		return Project{}, ErrConflict
	}
	current.Title = c.Title
	current.Body = c.Body
	current.Version++
	return current, nil
}

type Source struct{ ID, OwnerID string }

const MaxVersion int64 = 1<<53 - 1

var idPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }

type Item struct{ ID, IssueID, Title, State string }

type ItemPage struct {
	Items      []Item
	NextCursor string
}
