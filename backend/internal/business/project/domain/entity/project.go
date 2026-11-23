package entity

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Organization owners use the canonical UUID stored by the organization domain.
var organizationID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func IsOrganizationOwner(id string) bool { return organizationID.MatchString(id) }

var ErrInput = errors.New("invalid project input")

var ErrNotFound = errors.New("project not found")

var ErrForbidden = errors.New("only creator may edit")

var ErrConflict = errors.New("version conflict")

var ErrCreationConflict = errors.New("creation key conflict")

var ErrUnavailable = errors.New("project unavailable")

var ErrSourceUnavailable = errors.New("project source unavailable")

type (
	Field struct {
		ID, Name string
		Options  []string
	}
	View     struct{ ID, Name, FieldID string }
	Settings struct {
		Fields []Field
		Views  []View
	}
)

func (s Settings) Valid() bool {
	if len(s.Fields) > 32 || len(s.Views) > 32 {
		return false
	}
	label := func(name string) bool {
		return name != "" && name == strings.TrimSpace(name) && utf8.ValidString(name) && utf8.RuneCountInString(name) <= 80
	}
	fields, views := map[string]bool{}, map[string]bool{}
	for _, f := range s.Fields {
		if !ValidID(f.ID) || fields[f.ID] || !label(f.Name) || len(f.Options) > 32 {
			return false
		}
		fields[f.ID] = true
		options := map[string]bool{}
		for _, o := range f.Options {
			if options[o] || !label(o) {
				return false
			}
			options[o] = true
		}
	}
	for _, v := range s.Views {
		if !ValidID(v.ID) || views[v.ID] || !label(v.Name) || !fields[v.FieldID] {
			return false
		}
		views[v.ID] = true
	}
	return true
}

type Project struct {
	Settings                                                 Settings
	ID, DefaultRepositoryID, OwnerID, Title, Body, CreatorID string
	LinkedRepositoryIDs                                      []string
	SourceAvailable                                          bool
	Version                                                  int64
	CreatedAt, UpdatedAt                                     time.Time
	IssueCount, CompletedIssueCount                          int64
}

func (v Project) CanEdit(actor string) bool { return actor != "" && actor == v.CreatorID }

type Content struct{ Title, Body, SourceRepositoryID, CopyOwnerID, CopyProjectID string }

type Update struct {
	Settings *Settings
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
	if !current.CanEdit(actor) {
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
	settingsChanged := false
	if input.Settings != nil {
		if !input.Settings.Valid() {
			return Project{}, ErrInput
		}
		settingsChanged = !reflect.DeepEqual(*input.Settings, current.Settings)
		current.Settings = *input.Settings
	}
	if !settingsChanged && c.Title == current.Title && c.Body == current.Body {
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
