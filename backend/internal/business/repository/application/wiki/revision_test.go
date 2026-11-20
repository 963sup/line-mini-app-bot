package wiki

import (
	"testing"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
)

func TestRevisionAtVersionLimit(t *testing.T) {
	p := repositoryDomainEntityWiki.Page{AuthorID: "actor", Title: "title", Body: "body", Version: repositoryDomainEntity.MaxResourceNumber}
	if changed, err := p.NeedsRevision(p.Version, p.Title, p.Body); changed || err != nil {
		t.Fatal("unchanged content at the limit should succeed")
	}
	if _, err := p.NeedsRevision(p.Version, p.Title, "different"); err != repositoryDomainEntityWiki.ErrConflict {
		t.Fatal("new revision would exceed the limit")
	}
	if _, err := p.NeedsRevision(p.Version-1, p.Title, p.Body); err != repositoryDomainEntityWiki.ErrConflict {
		t.Fatal("stale version should conflict even for identical content")
	}
	if err := p.CheckCreationReplay("different actor", p.Title, p.Body); err != repositoryDomainEntityWiki.ErrConflict {
		t.Fatal("a different author cannot replay creation")
	}
}
