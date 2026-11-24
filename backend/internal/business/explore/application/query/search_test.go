package query

import (
	"context"
	"errors"
	"testing"
)

type sourceFunc func(context.Context, string, string) (Page, error)

func (f sourceFunc) ListVisible(ctx context.Context, actor, cursor string) (Page, error) {
	return f(ctx, actor, cursor)
}

func TestSearchFiltersOneVisiblePageAndScopesCursor(t *testing.T) {
	service := Service{Repositories: sourceFunc(func(_ context.Context, actor, cursor string) (Page, error) {
		if actor != "actor" || cursor != "position" {
			t.Fatalf("unexpected source scope: %q %q", actor, cursor)
		}
		return Page{
			Items:      []Item{{ID: "1", DisplayName: "Other"}, {ID: "2", DisplayName: "Alpha Repository"}},
			NextCursor: "next",
		}, nil
	})}

	first := Service{Repositories: sourceFunc(func(context.Context, string, string) (Page, error) {
		return Page{NextCursor: "position"}, nil
	})}
	page, err := first.Search(context.Background(), "actor", "repository", "alpha", "")
	if err != nil || page.NextCursor == "" {
		t.Fatalf("first page = %#v, %v", page, err)
	}
	page, err = service.Search(context.Background(), "actor", "repository", " ALPHA ", page.NextCursor)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "2" || page.NextCursor == "" {
		t.Fatalf("filtered page = %#v, %v", page, err)
	}
	if _, err := service.Search(context.Background(), "other", "repository", "alpha", page.NextCursor); !errors.Is(err, ErrInput) {
		t.Fatalf("actor-mismatched cursor error = %v", err)
	}
}

func TestSearchRejectsInvalidInputAndUnavailableSource(t *testing.T) {
	service := Service{}
	for _, test := range []struct{ actor, kind, name string }{
		{"", "repository", "name"},
		{"actor", "organization", "name"},
		{"actor", "repository", "  "},
	} {
		if _, err := service.Search(context.Background(), test.actor, test.kind, test.name, ""); !errors.Is(err, ErrInput) {
			t.Fatalf("Search(%q, %q, %q) error = %v", test.actor, test.kind, test.name, err)
		}
	}
	if _, err := service.Search(context.Background(), "actor", "repository", "name", ""); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("missing source error = %v", err)
	}
}
