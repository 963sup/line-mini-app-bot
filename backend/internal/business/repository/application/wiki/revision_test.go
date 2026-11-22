package wiki

import (
	"testing"
	"time"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/wiki"
)

func TestRevisionAtVersionLimit(t *testing.T) {
	p := repositoryDomainEntityWiki.Page{AuthorID: "actor", Title: "title", Body: "body", Version: repositoryDomainEntity.MaxResourceNumber}
	if _, changed, err := p.Revise(p.Version, p.AuthorID, p.Title, p.Body, time.Time{}); changed || err != nil {
		t.Fatal("unchanged content at the limit should succeed")
	}
	if _, _, err := p.Revise(p.Version, p.AuthorID, p.Title, "different", time.Time{}); err != repositoryDomainEntityWiki.ErrLimit {
		t.Fatal("new revision would exceed the limit")
	}
	if _, _, err := p.Revise(p.Version-1, p.AuthorID, p.Title, p.Body, time.Time{}); err != repositoryDomainEntityWiki.ErrVersion {
		t.Fatal("stale version should conflict even for identical content")
	}
	if err := p.CheckCreationReplay("different actor", p.Title, p.Body); err != repositoryDomainEntityWiki.ErrReplay {
		t.Fatal("a different author cannot replay creation")
	}
}
