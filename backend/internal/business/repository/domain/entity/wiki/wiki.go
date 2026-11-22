package wiki

import (
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

var ErrInput = errors.New("invalid wiki input")

var ErrConflict = errors.New("wiki title, version or creation key conflicts")

type Page struct {
	ID        string
	Title     string
	Body      string
	AuthorID  string
	Version   int64
	UpdatedAt time.Time
}

func Validate(title, body string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > 120 || strings.IndexFunc(title, unicode.IsControl) >= 0 || utf8.RuneCountInString(body) > 32000 {
		return "", ErrInput
	}
	return title, nil
}

// CheckCreationReplay compares a retried create with the immutable first revision.
func (p Page) CheckCreationReplay(actor, title, body string) error {
	if p.AuthorID != actor || p.Title != title || p.Body != body {
		return ErrConflict
	}
	return nil
}

// Revise returns the next page only when the requested edit changes its content.
func (p Page) Revise(expected int64, actor, title, body string, now time.Time) (Page, bool, error) {
	if expected != p.Version {
		return p, false, ErrConflict
	}
	if title == p.Title && body == p.Body {
		return p, false, nil
	}
	if p.Version >= repositoryDomainEntity.MaxResourceNumber {
		return p, false, ErrConflict
	}
	p.Title, p.Body, p.AuthorID = title, body, actor
	p.Version++
	p.UpdatedAt = now
	return p, true, nil
}
