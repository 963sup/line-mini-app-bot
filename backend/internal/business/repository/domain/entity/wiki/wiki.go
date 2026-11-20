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
	ID        string    `firestore:"id"`
	Title     string    `firestore:"title"`
	Body      string    `firestore:"body"`
	AuthorID  string    `firestore:"authorId"`
	Version   int64     `firestore:"version"`
	UpdatedAt time.Time `firestore:"updatedAt"`
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

func (p Page) NeedsRevision(expected int64, title, body string) (bool, error) {
	if expected != p.Version {
		return false, ErrConflict
	}
	if title == p.Title && body == p.Body {
		return false, nil
	}
	if p.Version >= repositoryDomainEntity.MaxResourceNumber {
		return false, ErrConflict
	}
	return true, nil
}
