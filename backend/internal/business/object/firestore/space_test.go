package firestore

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/object"
)

func TestSpaceCreationHashKeepsLegacyAndBindsAssignment(t *testing.T) {
	c := testContent()
	old := creationHash(c)
	c.SpaceID = strings.Repeat("a", 32)
	if creationHash(c) == old {
		t.Fatal("space assignment absent from creation intent")
	}
	assigned := creationHash(c)
	c.SpaceID = strings.Repeat("b", 32)
	if creationHash(c) == assigned {
		t.Fatal("different space shares intent")
	}
	c.SpaceID = ""
	if creationHash(c) != old {
		t.Fatal("unassigned legacy hash changed")
	}
}

func TestEmulatorRepositorySpaceAssignment(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	spaceID := strings.Repeat("d", 32)
	other := strings.Repeat("e", 32)
	now := time.Now().UTC()
	doc := map[string]any{"schemaVersion": 1, "id": spaceID, "repositoryId": r.ID, "name": "Site A", "creatorId": "actor-a", "version": 1, "createdAt": now, "updatedAt": now}
	// An equal ID in another repository grants no assignment authority.
	if _, err := store.repositoryRef(other).Collection("spaces").Doc(spaceID).Set(ctx, doc); err != nil {
		t.Fatal(err)
	}
	otherChannel := &Store{Client: store.Client, ChannelID: "another-channel"}
	if _, err := otherChannel.repositoryRef(r.ID).Collection("spaces").Doc(spaceID).Set(ctx, doc); err != nil {
		t.Fatal(err)
	}
	c := testContent()
	c.SpaceID = spaceID
	if _, err := store.Create(ctx, r, "actor-a", "cross-space-key-001", c); !errors.Is(err, object.ErrInput) {
		t.Fatalf("cross-repository space accepted: %v", err)
	}
	if collectionCount(t, ctx, store.repositoryRef(r.ID), "objects") != 0 || collectionCount(t, ctx, store.repositoryRef(r.ID), "objectCreationReceipts") != 0 {
		t.Fatal("rejected binding wrote an object or receipt")
	}
	if _, err := store.repositoryRef(r.ID).Collection("spaces").Doc(spaceID).Set(ctx, doc); err != nil {
		t.Fatal(err)
	}
	first, err := store.Create(ctx, r, "actor-a", "same-space-key-0001", c)
	if err != nil || first.SpaceID != spaceID {
		t.Fatal(first, err)
	}
	replay, err := store.Create(ctx, r, "actor-a", "same-space-key-0001", c)
	if err != nil || replay.ID != first.ID {
		t.Fatal("reply-lost retry", replay, err)
	}
	c.SpaceID = ""
	if _, err := store.Create(ctx, r, "actor-a", "same-space-key-0001", c); !errors.Is(err, object.ErrCreationConflict) {
		t.Fatal("space intent collision", err)
	}
	u := object.Update{Content: c, ExpectedVersion: 1}
	got, err := store.Update(ctx, r, "actor-a", first.ID, u)
	if err != nil || got.SpaceID != spaceID || got.Version != 1 {
		t.Fatal("omission not preserved", got, err)
	}
	unknown := strings.Repeat("f", 32)
	u.SpaceID = &unknown
	if _, err := store.Update(ctx, r, "actor-a", first.ID, u); !errors.Is(err, object.ErrInput) {
		t.Fatal("unknown space", err)
	}
	empty := ""
	u.SpaceID = &empty
	got, err = store.Update(ctx, r, "actor-a", first.ID, u)
	if err != nil || got.SpaceID != "" || got.Version != 2 {
		t.Fatal("clear", got, err)
	}
	snap, err := store.repositoryRef(r.ID).Collection("objects").Doc(first.ID).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := snap.Data()["spaceId"]; exists {
		t.Fatal("unassigned space field should be absent")
	}
}
