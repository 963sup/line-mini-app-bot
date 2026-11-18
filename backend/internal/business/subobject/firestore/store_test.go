package firestore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/object"
	objectdb "github.com/963sup/line-mini-app-bot/internal/business/object/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/subobject"
)

type repos struct{ r repository.Repository }

func (r repos) Get(_ context.Context, actor, id string) (repository.Repository, error) {
	if (actor != "owner" && actor != "member") || r.r.ID != id {
		return repository.Repository{}, repository.ErrNotFound
	}
	return r.r, nil
}

func setup(t *testing.T) (context.Context, *Store, *feature.Service, repository.Repository, object.Object) {
	t.Helper()
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("set loopback FIRESTORE_EMULATOR_HOST")
	}
	hostname, _, e := net.SplitHostPort(host)
	if e != nil || (hostname != "localhost" && !net.ParseIP(hostname).IsLoopback()) {
		t.Fatal("loopback emulator only")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	var suffix [8]byte
	if _, e = rand.Read(suffix[:]); e != nil {
		t.Fatal(e)
	}
	client, e := db.NewClient(ctx, "subobjects-"+hex.EncodeToString(suffix[:]))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = client.Close() })
	store := &Store{Client: client, ChannelID: "subobject-tests"}
	r := repository.Repository{ID: strings.Repeat("a", 32), GroupID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}
	if _, e = store.repository(r.ID).Set(ctx, map[string]any{"id": r.ID, "groupId": r.GroupID, "status": r.Status}); e != nil {
		t.Fatal(e)
	}
	objects := &object.Service{Repositories: repos{r}, Store: &objectdb.Store{Client: client, ChannelID: store.ChannelID}}
	parent, e := objects.Create(ctx, "owner", r.ID, "parent-create-key", object.Content{Name: "Fiber", Type: "non_work", Quantity: "140", Unit: "SET"})
	if e != nil {
		t.Fatal(e)
	}
	return ctx, store, &feature.Service{Repositories: repos{r}, Store: store}, r, parent
}

func TestEmulatorSubobjectAuthorityReceiptsAndParentIsolation(t *testing.T) {
	ctx, store, s, r, parent := setup(t)
	parentRef := store.parent(r.ID, parent.ID)
	before, e := parentRef.Get(ctx)
	if e != nil {
		t.Fatal(e)
	}
	c := feature.Content{Name: "Lay", Type: "work", Quantity: "1000.000001", Unit: "M"}
	const callers = 4
	results := make([]feature.Subobject, callers)
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Go(func() { results[i], errs[i] = s.Create(ctx, "owner", r.ID, parent.ID, "stable-child-key", c) })
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil || results[i].ID != results[0].ID || results[i].CompletedQuantity != "0" {
			t.Fatal(results[i], e)
		}
	}
	child := results[0]
	docs, e := parentRef.Collection("subobjects").Documents(ctx).GetAll()
	if e != nil || len(docs) != 1 {
		t.Fatal("concurrent duplicate", len(docs), e)
	}
	changed := c
	changed.Unit = "SET"
	if _, e = s.Create(ctx, "owner", r.ID, parent.ID, "stable-child-key", changed); !errors.Is(e, feature.ErrCreationConflict) {
		t.Fatal(e)
	}
	if _, e = s.Create(ctx, "member", r.ID, parent.ID, "member-child-key", c); !errors.Is(e, feature.ErrForbidden) {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, "member", r.ID, parent.ID, child.ID); e != nil {
		t.Fatal("member read", e)
	}
	if _, e = s.Get(ctx, "outsider", r.ID, parent.ID, child.ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, "owner", r.ID, strings.Repeat("c", 32), child.ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal("cross parent", e)
	}
	if _, e = s.Get(ctx, "owner", strings.Repeat("d", 32), parent.ID, child.ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal("cross repository", e)
	}
	done := "40.000001"
	u := feature.Update{Content: child.Content, CompletedQuantity: &done, ExpectedVersion: 1}
	if _, e = s.Update(ctx, "member", r.ID, parent.ID, child.ID, u); !errors.Is(e, feature.ErrForbidden) {
		t.Fatal(e)
	}
	changedRows := make([]feature.Subobject, 2)
	updateErrors := make([]error, 2)
	for i := range 2 {
		wg.Go(func() { changedRows[i], updateErrors[i] = s.Update(ctx, "owner", r.ID, parent.ID, child.ID, u) })
	}
	wg.Wait()
	success, conflicts := 0, 0
	for _, e := range updateErrors {
		if e == nil {
			success++
		} else if errors.Is(e, feature.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatal("version race", success, conflicts)
	}
	current, e := s.Get(ctx, "owner", r.ID, parent.ID, child.ID)
	if e != nil || current.CompletedQuantity != done {
		t.Fatal(current, e)
	}
	u.ExpectedVersion = 2
	noop, e := s.Update(ctx, "owner", r.ID, parent.ID, child.ID, u)
	if e != nil || noop.Version != 2 {
		t.Fatal(noop, e)
	}
	after, e := parentRef.Get(ctx)
	if e != nil || !reflect.DeepEqual(before.Data(), after.Data()) || !before.UpdateTime.Equal(after.UpdateTime) {
		t.Fatal("parent changed", e)
	}
	if _, e = store.repository(r.ID).Update(ctx, []db.Update{{Path: "status", Value: "ARCHIVED"}}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Update(ctx, "owner", r.ID, parent.ID, child.ID, u); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal("inactive", e)
	}
}

func TestEmulatorSubobjectPaginationScope(t *testing.T) {
	ctx, store, s, r, parent := setup(t)
	for i := range 21 {
		_, e := s.Create(ctx, "owner", r.ID, parent.ID, "pagination-key-"+strings.Repeat("x", i+1), feature.Content{Name: "Install", Type: "work", Quantity: "1", Unit: "M"})
		if e != nil {
			t.Fatal(e)
		}
	}
	page, e := s.List(ctx, "owner", r.ID, parent.ID, "")
	if e != nil || len(page.Items) != 20 || page.NextCursor == "" {
		t.Fatal(page, e)
	}
	last, e := s.List(ctx, "owner", r.ID, parent.ID, page.NextCursor)
	if e != nil || len(last.Items) != 1 || last.NextCursor != "" {
		t.Fatal(last, e)
	}
	if _, e = s.List(ctx, "member", r.ID, parent.ID, page.NextCursor); !errors.Is(e, feature.ErrInput) {
		t.Fatal("actor cursor", e)
	}
	if _, e = s.List(ctx, "owner", r.ID, strings.Repeat("c", 32), page.NextCursor); !errors.Is(e, feature.ErrInput) {
		t.Fatal("parent cursor", e)
	}
	if _, e = store.parent(r.ID, parent.ID).Update(ctx, []db.Update{{Path: "creatorId", Value: "member"}}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Create(ctx, "owner", r.ID, parent.ID, "stale-owner-key-01", feature.Content{Name: "new", Type: "non_work", Quantity: "1", Unit: "SET"}); !errors.Is(e, feature.ErrForbidden) {
		t.Fatal("current parent owner", e)
	}
}
