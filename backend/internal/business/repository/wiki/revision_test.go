package wiki

import (
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

func TestRevisionAtVersionLimit(t *testing.T) {
	p := Page{AuthorID: "actor", Title: "title", Body: "body", Version: repository.MaxResourceNumber}
	if changed, err := p.NeedsRevision(p.Version, p.Title, p.Body); changed || err != nil {
		t.Fatal("unchanged content at the limit should succeed")
	}
	if _, err := p.NeedsRevision(p.Version, p.Title, "different"); err != ErrConflict {
		t.Fatal("new revision would exceed the limit")
	}
	if _, err := p.NeedsRevision(p.Version-1, p.Title, p.Body); err != ErrConflict {
		t.Fatal("stale version should conflict even for identical content")
	}
	if err := p.CheckCreationReplay("different actor", p.Title, p.Body); err != ErrConflict {
		t.Fatal("a different author cannot replay creation")
	}
}
