package query

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInput = errors.New("invalid explore query")
var ErrUnavailable = errors.New("explore source unavailable")

type Item struct {
	ID, DisplayName string
}

type Page struct {
	Items      []Item
	NextCursor string
}

// RepositorySource returns one ID-ordered page of active, opted-in repositories.
type RepositorySource interface {
	ListVisible(context.Context, string, string) (Page, error)
}

type Service struct{ Repositories RepositorySource }

func (s Service) Search(ctx context.Context, actor, kind, name, cursor string) (Page, error) {
	query := strings.ToLower(strings.TrimSpace(name))
	if actor == "" || kind != "repository" || query == "" || utf8.RuneCountInString(name) > 128 || len(cursor) > 1024 {
		return Page{}, ErrInput
	}
	scopeBytes, _ := json.Marshal([]string{"explore-v1", actor, kind, query})
	scope := sha256.Sum256(scopeBytes)
	prefix := base64.RawURLEncoding.EncodeToString(scope[:]) + "."
	position := ""
	if cursor != "" {
		var valid bool
		position, valid = strings.CutPrefix(cursor, prefix)
		if !valid || position == "" {
			return Page{}, ErrInput
		}
	}
	if s.Repositories == nil {
		return Page{}, ErrUnavailable
	}
	page, err := s.Repositories.ListVisible(ctx, actor, position)
	if err != nil {
		return Page{}, err
	}
	result := Page{Items: make([]Item, 0, len(page.Items))}
	for _, item := range page.Items {
		if strings.Contains(strings.ToLower(item.DisplayName), query) {
			result.Items = append(result.Items, item)
		}
	}
	if page.NextCursor != "" {
		result.NextCursor = prefix + page.NextCursor
	}
	return result, nil
}
