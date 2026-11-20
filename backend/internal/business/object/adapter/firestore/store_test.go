package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
)

// These tests require an explicitly selected loopback emulator, never ADC or a
// production project. Each fixture owns a random project and an expiring context.
func emulatorStore(t *testing.T) (context.Context, *Store, repositoryDomainEntity.Repository) {
	t.Helper()
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("set FIRESTORE_EMULATOR_HOST to run real Firestore transaction regressions")
	}
	hostname, _, err := net.SplitHostPort(host)
	if err != nil || (hostname != "localhost" && !net.ParseIP(hostname).IsLoopback()) {
		t.Fatal("transaction tests require a loopback emulator endpoint")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	client, err := db.NewClient(ctx, "objects-test-"+hex.EncodeToString(random[:]))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	store := &Store{Client: client, ChannelID: "isolated-object-tests"}
	r := repositoryDomainEntity.Repository{ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", OwnerID: "Cbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Status: "ACTIVE"}
	if _, err := store.repositoryRef(r.ID).Set(ctx, map[string]any{"id": r.ID, "ownerId": r.OwnerID, "status": r.Status}); err != nil {
		t.Fatal(err)
	}
	return ctx, store, r
}

func testContent() objectDomainEntity.Content {
	return objectDomainEntity.Content{Name: "Test engineering item", Type: "Work", Quantity: "11", Unit: "SET", SourceDocument: "test-document", SourceRevision: "v0", SourceItemNo: "10"}
}

func collectionCount(t *testing.T, ctx context.Context, parent *db.DocumentRef, name string) int {
	t.Helper()
	docs, err := parent.Collection(name).Documents(ctx).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	return len(docs)
}

func TestEmulatorConcurrentCreationAndReceiptReplay(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	content := testContent()
	const callers = 6
	results := make([]objectDomainEntity.Object, callers)
	errs := make([]error, callers)
	var group sync.WaitGroup
	start := make(chan struct{})
	for i := range callers {
		group.Go(func() {
			<-start
			results[i], errs[i] = store.Create(ctx, r, "actor-a", "same-creation-key-0001", content)
		})
	}
	close(start)
	group.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("caller %d: %v", i, err)
		}
		if results[i].ID != results[0].ID || results[i].Version != 1 || results[i].Status != "not_started" {
			t.Fatalf("duplicate create produced different resource: %#v", results[i])
		}
	}
	parent := store.repositoryRef(r.ID)
	if n := collectionCount(t, ctx, parent, "objects"); n != 1 {
		t.Fatalf("objects=%d, want 1", n)
	}
	if n := collectionCount(t, ctx, parent, "objectCreationReceipts"); n != 1 {
		t.Fatalf("receipts=%d, want 1", n)
	}
	if n := collectionCount(t, ctx, parent, "resourceNumbers"); n != 0 {
		t.Fatalf("unnumbered objects allocated resource numbers: %d", n)
	}
	changed := content
	changed.Name = "Different intent"
	if _, err := store.Create(ctx, r, "actor-a", "same-creation-key-0001", changed); !errors.Is(err, objectDomainEntity.ErrCreationConflict) {
		t.Fatalf("changed payload error=%v", err)
	}
	updated, err := store.Update(ctx, r, "actor-a", results[0].ID, objectDomainEntity.Update{Content: content, Status: "completed", ExpectedVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	replay, err := store.Create(ctx, r, "actor-a", "same-creation-key-0001", content)
	if err != nil || replay.ID != updated.ID || replay.Version != 2 || replay.Status != "completed" {
		t.Fatalf("replay must return current snapshot: %#v, %v", replay, err)
	}
	other := r
	other.ID = "cccccccccccccccccccccccccccccccc"
	if _, err := store.repositoryRef(other.ID).Set(ctx, map[string]any{"id": other.ID, "ownerId": other.OwnerID, "status": "ACTIVE"}); err != nil {
		t.Fatal(err)
	}
	independent, err := store.Create(ctx, other, "actor-a", "same-creation-key-0001", content)
	if err != nil || independent.ID == replay.ID {
		t.Fatalf("cross-repository receipt isolation: %#v %v", independent, err)
	}
}

func TestEmulatorContainerAndReceiptFailuresDoNotWrite(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	parent := store.repositoryRef(r.ID)
	for _, mutation := range []map[string]any{
		{"id": r.ID, "ownerId": r.OwnerID, "status": "INACTIVE"},
		{"id": r.ID, "ownerId": "Cdddddddddddddddddddddddddddddddd", "status": "ACTIVE"},
	} {
		if _, err := parent.Set(ctx, mutation); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Create(ctx, r, "actor-a", "rejected-creation-key", testContent()); !errors.Is(err, objectDomainEntity.ErrNotFound) {
			t.Fatalf("container error=%v", err)
		}
		if collectionCount(t, ctx, parent, "objects") != 0 || collectionCount(t, ctx, parent, "objectCreationReceipts") != 0 {
			t.Fatal("rejected creation left resource or receipt")
		}
	}
	if _, err := parent.Set(ctx, map[string]any{"id": r.ID, "ownerId": r.OwnerID, "status": "ACTIVE"}); err != nil {
		t.Fatal(err)
	}
	rec := parent.Collection("objectCreationReceipts").Doc(firestorekey.JSONDigest([]string{"actor-a", "corrupt-receipt-key"}))
	if _, err := rec.Set(ctx, map[string]any{"schemaVersion": 1, "id": "invalid", "payloadHash": "broken"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(ctx, r, "actor-a", "corrupt-receipt-key", testContent()); !errors.Is(err, objectDomainEntity.ErrUnavailable) {
		t.Fatalf("corrupt receipt error=%v", err)
	}
	if collectionCount(t, ctx, parent, "objects") != 0 || collectionCount(t, ctx, parent, "objectCreationReceipts") != 1 {
		t.Fatal("receipt failure committed a partial creation")
	}
}

func TestEmulatorUpdateChecksAuthorVersionAndPreservesNoOp(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	original, err := store.Create(ctx, r, "actor-a", "update-creation-key", testContent())
	if err != nil {
		t.Fatal(err)
	}
	input := objectDomainEntity.Update{Content: original.Content, Status: "in_progress", ExpectedVersion: 1}
	input.SourceDocument = ""
	input.SourceRevision = ""
	input.SourceItemNo = ""
	for _, tc := range []struct {
		actor   string
		version int64
		want    error
	}{{"actor-b", 1, objectDomainEntity.ErrForbidden}, {"actor-a", 2, objectDomainEntity.ErrConflict}} {
		input.ExpectedVersion = tc.version
		if _, err := store.Update(ctx, r, tc.actor, original.ID, input); !errors.Is(err, tc.want) {
			t.Fatalf("actor %s version %d error=%v", tc.actor, tc.version, err)
		}
	}
	input.ExpectedVersion = 1
	updated, err := store.Update(ctx, r, "actor-a", original.ID, input)
	if err != nil || updated.Version != 2 || updated.SourceDocument != original.SourceDocument || updated.SourceRevision != original.SourceRevision || updated.SourceItemNo != original.SourceItemNo {
		t.Fatalf("update: %#v %v", updated, err)
	}
	ref := store.repositoryRef(r.ID).Collection("objects").Doc(original.ID)
	if n := collectionCount(t, ctx, ref, "changes"); n != 1 {
		t.Fatalf("changes=%d", n)
	}
	input.ExpectedVersion = 2
	noOp, err := store.Update(ctx, r, "actor-a", original.ID, input)
	if err != nil || noOp.Version != 2 || !noOp.UpdatedAt.Equal(updated.UpdatedAt) || collectionCount(t, ctx, ref, "changes") != 1 {
		t.Fatalf("no-op wrote change: %#v %v", noOp, err)
	}
	input.Status = "completed"
	input.ExpectedVersion = 1
	if _, err := store.Update(ctx, r, "actor-a", original.ID, input); !errors.Is(err, objectDomainEntity.ErrConflict) {
		t.Fatalf("stale update error=%v", err)
	}
	stored, err := store.Get(ctx, r, original.ID)
	if err != nil || stored.Status != "in_progress" || stored.Version != 2 {
		t.Fatalf("failed update changed state: %#v %v", stored, err)
	}
	empty := ""
	input.ExpectedVersion = 2
	input.Sources.SourceDocument = &empty
	cleared, err := store.Update(ctx, r, "actor-a", original.ID, input)
	if err != nil || cleared.SourceDocument != "" || cleared.SourceRevision != original.SourceRevision || cleared.Version != 3 {
		t.Fatalf("explicit clear must preserve omitted provenance: %#v %v", cleared, err)
	}
}

func TestEmulatorListPaginatesWithoutMissingOrDuplicateObjects(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	if err := store.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		for n := 1; n <= 21; n++ {
			id := fmt.Sprintf("%032x", n)
			item := objectDomainEntity.Object{Content: testContent(), ID: id, RepositoryID: r.ID, CreatorID: "actor-a", Status: "not_started", Version: 1, CreatedAt: now, UpdatedAt: now}
			if err := tx.Set(store.repositoryRef(r.ID).Collection("objects").Doc(id), encode(item)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	first, err := store.List(ctx, r, "actor-a", "")
	if err != nil || len(first.Items) != 20 || first.NextCursor == "" {
		t.Fatalf("first page: %d items, cursor %q, error %v", len(first.Items), first.NextCursor, err)
	}
	second, err := store.List(ctx, r, "actor-a", first.NextCursor)
	if err != nil || len(second.Items) != 1 || second.NextCursor != "" {
		t.Fatalf("second page: %d items, cursor %q, error %v", len(second.Items), second.NextCursor, err)
	}
	all := append(first.Items, second.Items...)
	for i, item := range all {
		want := fmt.Sprintf("%032x", i+1)
		if item.ID != want || item.RepositoryID != r.ID || item.Content != testContent() {
			t.Fatalf("position %d: got %#v, want id %s", i, item, want)
		}
	}
	got, err := store.Get(ctx, r, all[0].ID)
	if err != nil || got.ID != all[0].ID || got.Content != all[0].Content {
		t.Fatalf("get: %#v, error %v", got, err)
	}
	if _, err := store.List(ctx, r, "actor-b", first.NextCursor); !errors.Is(err, objectDomainEntity.ErrInput) {
		t.Fatalf("cursor must reject another actor: %v", err)
	}
	other := r
	other.ID = "cccccccccccccccccccccccccccccccc"
	if _, err := store.List(ctx, other, "actor-a", first.NextCursor); !errors.Is(err, objectDomainEntity.ErrInput) {
		t.Fatalf("cursor must reject another repository: %v", err)
	}
	if _, err := store.Get(ctx, other, got.ID); !errors.Is(err, objectDomainEntity.ErrNotFound) {
		t.Fatalf("get must not leak another repository: %v", err)
	}
}
