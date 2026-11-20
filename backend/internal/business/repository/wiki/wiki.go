// Package wiki defines repository knowledge pages and their immutable revisions.
package wiki

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

var (
	ErrInput    = errors.New("invalid wiki input")
	ErrConflict = errors.New("wiki title, version or creation key conflicts")
)

type Page struct {
	ID        string    `firestore:"id"`
	Title     string    `firestore:"title"`
	Body      string    `firestore:"body"`
	AuthorID  string    `firestore:"authorId"`
	Version   int64     `firestore:"version"`
	UpdatedAt time.Time `firestore:"updatedAt"`
}
type Store interface {
	GetWiki(context.Context, repository.Repository, string) (Page, error)
	ListWiki(context.Context, repository.Repository, string, string, string) ([]Page, string, error)
	SaveWiki(context.Context, repository.Repository, string, string, string, int64, string, string) (Page, error)
}
type Service struct {
	Repositories *repository.Service
	Store        Store
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
	if p.Version >= repository.MaxResourceNumber {
		return false, ErrConflict
	}
	return true, nil
}
