package entity

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
)

import (
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInput            = errors.New("invalid discussion input")
	ErrNotFound         = errors.New("discussion or comment not found")
	ErrForbidden        = errors.New("only the author may edit")
	ErrConflict         = errors.New("discussion version conflict")
	ErrCreationConflict = errors.New("creation key conflict")
)

type Discussion struct {
	ID, RepositoryID, Title, Body, AuthorID string
	Number, Version                         int64
	CreatedAt, UpdatedAt                    time.Time
	LabelIDs                                []string
}
type Comment struct {
	ID, RepositoryID, DiscussionID, Body, AuthorID string
	Version                                        int64
	CreatedAt, UpdatedAt                           time.Time
}
type (
	Content struct{ Title, Body string }
	Update  struct {
		Content
		ExpectedVersion int64
		LabelIDs        *[]string
	}
)

func Normalize(c Content) (Content, error) {
	c.Title, c.Body = strings.TrimSpace(c.Title), strings.TrimSpace(c.Body)
	if !utf8.ValidString(c.Title) || !utf8.ValidString(c.Body) || c.Title == "" || utf8.RuneCountInString(c.Title) > 120 || utf8.RuneCountInString(c.Body) > 8000 {
		return Content{}, ErrInput
	}
	return c, nil
}

func NormalizeComment(body string) (string, error) {
	body = strings.TrimSpace(body)
	if !utf8.ValidString(body) || body == "" || utf8.RuneCountInString(body) > 4000 {
		return "", ErrInput
	}
	return body, nil
}

func ApplyUpdate(current Discussion, actor string, input Update) (Discussion, error) {
	if actor == "" || (actor != current.AuthorID && (input.Title != current.Title || input.Body != current.Body || input.LabelIDs == nil)) {
		return Discussion{}, ErrForbidden
	}
	c, err := Normalize(input.Content)
	if err != nil || input.ExpectedVersion < 1 {
		return Discussion{}, ErrInput
	}
	if current.Version != input.ExpectedVersion {
		return Discussion{}, ErrConflict
	}
	ids := current.LabelIDs
	if input.LabelIDs != nil {
		ids, err = repositoryDomainEntityLabels.NormalizeIDs(*input.LabelIDs)
		if err != nil {
			return Discussion{}, ErrInput
		}
	}
	if c.Title == current.Title && c.Body == current.Body && slices.Equal(current.LabelIDs, ids) {
		return current, nil
	}
	if current.Version >= repositoryDomainEntity.MaxResourceNumber {
		return Discussion{}, ErrConflict
	}
	current.Title, current.Body = c.Title, c.Body
	current.LabelIDs = append([]string{}, ids...)
	current.Version++
	return current, nil
}

func ApplyCommentUpdate(current Comment, actor, body string, version int64) (Comment, error) {
	if actor == "" || actor != current.AuthorID {
		return Comment{}, ErrForbidden
	}
	body, err := NormalizeComment(body)
	if err != nil || version < 1 {
		return Comment{}, ErrInput
	}
	if current.Version != version {
		return Comment{}, ErrConflict
	}
	if body == current.Body {
		return current, nil
	}
	if current.Version >= repositoryDomainEntity.MaxResourceNumber {
		return Comment{}, ErrConflict
	}
	current.Body = body
	current.Version++
	return current, nil
}
