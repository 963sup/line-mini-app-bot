package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	objectApplication "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

import (
	"context"
	"errors"
	"strings"
	"testing"

	db "cloud.google.com/go/firestore"
)

type progressRepos struct {
	r repositoryDomainEntity.Repository
}

func (p progressRepos) Get(_ context.Context, actor, rid string) (repositoryDomainEntity.Repository, error) {
	if actor != "actor-a" && actor != "actor-b" || rid != p.r.ID {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	return p.r, nil
}

func TestEmulatorExactProgressAndLegacyDocuments(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	service := &objectApplication.Service{Repositories: progressRepos{r}, Store: store}
	c := testContent()
	c.Type = objectDomainEntity.TypeWork
	c.Quantity = "999999999999999999.999999"
	j, e := service.Create(ctx, "actor-a", r.ID, "progress-create-key", c)
	if e != nil || j.CompletedQuantity != "0" || j.Status != objectDomainEntity.StatusNotStarted {
		t.Fatal(j, e)
	}
	replay, e := service.Create(ctx, "actor-a", r.ID, "progress-create-key", c)
	if e != nil || j.ID != replay.ID {
		t.Fatal(e)
	}
	done := "999999999999999999.999998"
	u := objectDomainEntity.Update{Content: j.Content, CompletedQuantity: &done, ExpectedVersion: 1}
	updated, e := service.Update(ctx, "actor-a", r.ID, j.ID, u)
	if e != nil || updated.Status != objectDomainEntity.StatusInProgress || updated.Version != 2 {
		t.Fatal(updated, e)
	}
	if _, e = service.Update(ctx, "actor-b", r.ID, j.ID, u); !errors.Is(e, objectDomainEntity.ErrForbidden) {
		t.Fatal("author", e)
	}
	if _, e = service.Update(ctx, "actor-a", r.ID, j.ID, u); !errors.Is(e, objectDomainEntity.ErrConflict) {
		t.Fatal("version", e)
	}
	u.ExpectedVersion = 2
	noOp, e := service.Update(ctx, "actor-a", r.ID, j.ID, u)
	if e != nil || noOp.Version != 2 {
		t.Fatal("absolute retry", noOp, e)
	}
	u.Type = objectDomainEntity.TypeNonWork
	u.CompletedQuantity = nil
	if _, e = service.Update(ctx, "actor-a", r.ID, j.ID, u); !errors.Is(e, objectDomainEntity.ErrProgressRemoval) {
		t.Fatal(e)
	}
	snap, e := store.repositoryRef(r.ID).Collection("objects").Doc(j.ID).Get(ctx)
	if e != nil || snap.Data()["completedQuantity"] != done {
		t.Fatal("stored exact value", e)
	}
	// Recreate a historical document+receipt with no new field and the original v1 hash.
	legacy := testContent()
	legacy.Type = objectDomainEntity.TypeWork
	old, e := store.Create(ctx, r, "actor-a", "legacy-work-key-001", legacy)
	if e != nil {
		t.Fatal(e)
	}
	replay, e = service.Create(ctx, "actor-a", r.ID, "legacy-work-key-001", legacy)
	if e != nil || replay.ID != old.ID || replay.CompletedQuantity != "" {
		t.Fatal("legacy receipt", replay, e)
	}
	read, e := service.Get(ctx, "actor-a", r.ID, old.ID)
	if e != nil || read.CompletedQuantity != "" {
		t.Fatal("legacy read", read, e)
	}
	patch := objectDomainEntity.Update{Content: legacy, Status: objectDomainEntity.StatusCompleted, ExpectedVersion: 1}
	read, e = service.Update(ctx, "actor-a", r.ID, old.ID, patch)
	if e != nil || read.Status != objectDomainEntity.StatusCompleted || read.CompletedQuantity != "" {
		t.Fatal("legacy patch", read, e)
	}
	snapshot, e := store.repositoryRef(r.ID).Collection("objects").Doc(old.ID).Get(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, exists := snapshot.Data()["completedQuantity"]; exists {
		t.Fatal("historical absence changed")
	}
	c.Type = objectDomainEntity.TypeNonWork
	non, e := service.Create(ctx, "actor-a", r.ID, "non-work-create-key", c)
	if e != nil || non.CompletedQuantity != "" {
		t.Fatal(non, e)
	}
	if _, e = service.Get(ctx, "outsider", r.ID, j.ID); !errors.Is(e, objectDomainEntity.ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = service.Get(ctx, "actor-a", strings.Repeat("c", 32), j.ID); !errors.Is(e, objectDomainEntity.ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = store.repositoryRef(r.ID).Update(ctx, []db.Update{{Path: "status", Value: "ARCHIVED"}}); e != nil {
		t.Fatal(e)
	}
	if _, e = service.Update(ctx, "actor-a", r.ID, j.ID, u); !errors.Is(e, objectDomainEntity.ErrNotFound) {
		t.Fatal("active container", e)
	}
}
