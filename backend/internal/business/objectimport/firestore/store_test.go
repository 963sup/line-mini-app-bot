package firestore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/object"
	objectdb "github.com/963sup/line-mini-app-bot/internal/business/object/firestore"
	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

func importEmulator(t *testing.T) (context.Context, *db.Client, repository.Repository, *db.DocumentRef) {
	t.Helper()
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("set FIRESTORE_EMULATOR_HOST to run real import transaction regressions")
	}
	hostname, _, err := net.SplitHostPort(host)
	if err != nil || (hostname != "localhost" && !net.ParseIP(hostname).IsLoopback()) {
		t.Fatal("import tests require a loopback emulator endpoint")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	client, err := db.NewClient(ctx, "imports-test-"+hex.EncodeToString(random[:]))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	r := repository.Repository{ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", GroupID: "Cbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Status: "ACTIVE"}
	parent := client.Collection("channels").Doc("isolated-import-tests").Collection("repositories").Doc(r.ID)
	if _, err := parent.Set(ctx, map[string]any{"id": r.ID, "groupId": r.GroupID, "status": "ACTIVE"}); err != nil {
		t.Fatal(err)
	}
	return ctx, client, r, parent
}

// Only external membership observation is substituted; both draft and Object
// writes use the real SDK and emulator transaction implementation.
type importRepositories struct{ repository repository.Repository }

func (r importRepositories) Get(_ context.Context, actor, id string) (repository.Repository, error) {
	if (actor != "actor-a" && actor != "actor-b") || id != r.repository.ID {
		return repository.Repository{}, repository.ErrNotFound
	}
	return r.repository, nil
}

func importServices(client *db.Client, r repository.Repository) (*feature.Service, *Store, *object.Service) {
	repositories := importRepositories{r}
	objects := &object.Service{Repositories: repositories, Store: &objectdb.Store{Client: client, ChannelID: "isolated-import-tests"}}
	store := &Store{Client: client, ChannelID: "isolated-import-tests"}
	return &feature.Service{Repositories: repositories, Objects: objects, Store: store}, store, objects
}

func importCreation() feature.Creation {
	return feature.Creation{SourceDocument: "sanitized-document", SourceRevision: "v0", Pages: []feature.Page{{Number: 2, Text: "Original whole-page-only sentinel. Item10 pump eleven SET. Item20 labor two SET."}}, Rows: []feature.Candidate{
		{Fields: feature.Fields{Name: "Pump", Description: "Scope A", Type: "Equipment", Quantity: "11", Unit: "SET"}, SourceItemNo: "10", Evidence: []feature.Evidence{{Page: 2, Quote: "Item10 pump eleven SET."}}},
		{Fields: feature.Fields{Name: "Labor", Description: "Scope B", Type: "Work", Quantity: "2", Unit: "SET"}, SourceItemNo: "20", Evidence: []feature.Evidence{{Page: 2, Quote: "Item20 labor two SET."}}},
	}}
}

func importReview(j feature.Job) feature.Review {
	input := feature.Review{ExpectedVersion: j.Version, Rows: []feature.ReviewedRow{}}
	for _, row := range j.Rows {
		input.Rows = append(input.Rows, feature.ReviewedRow{ID: row.ID, Fields: row.Fields})
	}
	return input
}

func importCount(t *testing.T, ctx context.Context, parent *db.DocumentRef, collection string) int {
	t.Helper()
	docs, err := parent.Collection(collection).Documents(ctx).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	return len(docs)
}

func TestEmulatorDraftCreationReceiptIsAtomicAndScoped(t *testing.T) {
	ctx, client, r, parent := importEmulator(t)
	service, _, _ := importServices(client, r)
	const callers = 4
	results := make([]feature.Job, callers)
	errs := make([]error, callers)
	var group sync.WaitGroup
	start := make(chan struct{})
	for n := range callers {
		group.Go(func() {
			<-start
			results[n], errs[n] = service.Create(ctx, "actor-a", r.ID, "same-draft-key-0001", importCreation())
		})
	}
	close(start)
	group.Wait()
	for n, err := range errs {
		if err != nil {
			t.Fatalf("caller%d: %v", n, err)
		}
		if results[n].ID != results[0].ID || results[n].Rows[0].ID != results[0].Rows[0].ID || results[n].Status != "draft" || results[n].Version != 1 {
			t.Fatalf("non-idempotent draft: %#v", results[n])
		}
	}
	if importCount(t, ctx, parent, "objectImports") != 1 || importCount(t, ctx, parent, "objectImportCreationReceipts") != 1 || importCount(t, ctx, parent, "objects") != 0 {
		t.Fatal("draft creation must atomically save one draft/receipt and no Objects")
	}
	snap, err := parent.Collection("objectImports").Doc(results[0].ID).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(fmt.Sprint(snap.Data()), "whole-page-only sentinel") {
		t.Fatal("raw whole page retained in draft")
	}
	changed := importCreation()
	changed.Rows[0].Name = "Changed proposal"
	if _, err := service.Create(ctx, "actor-a", r.ID, "same-draft-key-0001", changed); !errors.Is(err, feature.ErrCreationConflict) {
		t.Fatalf("payload conflict: %v", err)
	}
	otherActor, err := service.Create(ctx, "actor-b", r.ID, "same-draft-key-0001", importCreation())
	if err != nil || otherActor.ID == results[0].ID {
		t.Fatalf("actor receipt isolation: %#v %v", otherActor, err)
	}
	if _, err := service.Get(ctx, "actor-b", r.ID, results[0].ID); !errors.Is(err, feature.ErrNotFound) {
		t.Fatalf("other member must not read draft: %v", err)
	}
	if _, err := service.Get(ctx, "actor-a", "cccccccccccccccccccccccccccccccc", results[0].ID); !errors.Is(err, feature.ErrNotFound) {
		t.Fatalf("cross-repository draft: %v", err)
	}
}

func TestEmulatorReviewFreezesExactVersionBeforeObjects(t *testing.T) {
	ctx, client, r, parent := importEmulator(t)
	service, store, _ := importServices(client, r)
	draft, err := service.Create(ctx, "actor-a", r.ID, "review-draft-key-01", importCreation())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Confirm(ctx, "actor-a", r.ID, draft.ID, draft.Version); !errors.Is(err, feature.ErrConflict) {
		t.Fatalf("unreviewed confirm: %v", err)
	}
	review := importReview(draft)
	if _, err := service.Review(ctx, "actor-b", r.ID, draft.ID, review); !errors.Is(err, feature.ErrNotFound) {
		t.Fatalf("other member review: %v", err)
	}
	reviewed, err := service.Review(ctx, "actor-a", r.ID, draft.ID, review)
	if err != nil || reviewed.Status != "reviewed" || reviewed.Version != 2 {
		t.Fatalf("review: %#v %v", reviewed, err)
	}
	if _, err := service.Review(ctx, "actor-a", r.ID, draft.ID, review); !errors.Is(err, feature.ErrConflict) {
		t.Fatalf("stale review: %v", err)
	}
	frozen, err := store.BeginConfirm(ctx, r, "actor-a", draft.ID, reviewed.Version)
	if err != nil || frozen.Status != "confirming" || frozen.ConfirmedFromReviewVersion != 2 {
		t.Fatalf("freeze: %#v %v", frozen, err)
	}
	if _, err := service.Review(ctx, "actor-a", r.ID, draft.ID, importReview(frozen)); !errors.Is(err, feature.ErrConflict) {
		t.Fatalf("frozen review: %v", err)
	}
	if _, err := store.BeginConfirm(ctx, r, "actor-a", draft.ID, frozen.Version); !errors.Is(err, feature.ErrConflict) {
		t.Fatalf("retry must use original review version: %v", err)
	}
	retry, err := store.BeginConfirm(ctx, r, "actor-a", draft.ID, reviewed.Version)
	if err != nil || retry.Version != frozen.Version || retry.ConfirmedFromReviewVersion != 2 {
		t.Fatalf("frozen retry changed snapshot: %#v %v", retry, err)
	}
	if importCount(t, ctx, parent, "objects") != 0 {
		t.Fatal("freeze must not directly write Objects")
	}
}

type unknownAfterCommit struct {
	real  *object.Service
	calls atomic.Int64
}

func (w *unknownAfterCommit) Create(ctx context.Context, actor, rid, key string, content object.Content) (object.Object, error) {
	result, err := w.real.Create(ctx, actor, rid, key, content)
	if err == nil && w.calls.Add(1) == 2 {
		return object.Object{}, object.ErrUnavailable
	}
	return result, err
}

func TestEmulatorConfirmRecoversCommittedUnknownAndConcurrentRetry(t *testing.T) {
	ctx, client, r, parent := importEmulator(t)
	service, _, objects := importServices(client, r)
	draft, err := service.Create(ctx, "actor-a", r.ID, "confirm-draft-key-1", importCreation())
	if err != nil {
		t.Fatal(err)
	}
	reviewed, err := service.Review(ctx, "actor-a", r.ID, draft.ID, importReview(draft))
	if err != nil {
		t.Fatal(err)
	}
	service.Objects = &unknownAfterCommit{real: objects}
	partial, err := service.Confirm(ctx, "actor-a", r.ID, draft.ID, reviewed.Version)
	if !errors.Is(err, feature.ErrUnavailable) || partial.Status != "confirming" || partial.Rows[0].ResultObjectID == "" || partial.Rows[1].ResultObjectID != "" {
		t.Fatalf("partial response: %#v %v", partial, err)
	}
	if importCount(t, ctx, parent, "objects") != 2 {
		t.Fatal("injected unknown must occur after second real Object commit")
	}
	const callers = 3
	results := make([]feature.Job, callers)
	errs := make([]error, callers)
	var group sync.WaitGroup
	start := make(chan struct{})
	for n := range callers {
		group.Go(func() {
			<-start
			results[n], errs[n] = service.Confirm(ctx, "actor-a", r.ID, draft.ID, reviewed.Version)
		})
	}
	close(start)
	group.Wait()
	for n, err := range errs {
		if err != nil {
			t.Fatalf("confirm caller%d: %v", n, err)
		}
		if results[n].Status != "completed" || results[n].ConfirmedFromReviewVersion != reviewed.Version {
			t.Fatalf("result: %#v", results[n])
		}
		for i, row := range results[n].Rows {
			if row.ResultObjectID == "" || row.ResultObjectID != results[0].Rows[i].ResultObjectID {
				t.Fatal("concurrent result mapping diverged")
			}
		}
	}
	if results[0].Rows[0].ResultObjectID == results[0].Rows[1].ResultObjectID || importCount(t, ctx, parent, "objects") != 2 || importCount(t, ctx, parent, "objectCreationReceipts") != 2 {
		t.Fatal("retry duplicated or merged distinct rows")
	}
	for _, row := range results[0].Rows {
		got, err := objects.Get(ctx, "actor-a", r.ID, row.ResultObjectID)
		if err != nil || got.Name != row.Name || got.Quantity != row.Quantity || got.SourceDocument != reviewed.SourceDocument || got.SourceItemNo != row.SourceItemNo || got.Status != "not_started" {
			t.Fatalf("mapped Object mismatch: %#v %v", got, err)
		}
	}
	if _, err := service.Review(ctx, "actor-a", r.ID, draft.ID, importReview(results[0])); !errors.Is(err, feature.ErrConflict) {
		t.Fatalf("completed snapshot mutable: %v", err)
	}
}
