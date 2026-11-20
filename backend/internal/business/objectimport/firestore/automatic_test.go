package firestore

import (
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"

	feature "github.com/963sup/line-mini-app-bot/internal/business/objectimport"
	"github.com/963sup/line-mini-app-bot/internal/business/objectimport/automatic"
)

type autoProducer struct {
	calls   atomic.Int32
	barrier *sync.WaitGroup
}

func (p *autoProducer) Rows(_ context.Context, s feature.Source) ([]automatic.Candidate, error) {
	n := p.calls.Add(1)
	if p.barrier != nil {
		p.barrier.Done()
		p.barrier.Wait()
	}
	return []automatic.Candidate{{Name: "Pump variant " + string(rune('A'+n)), Quantity: "11.00", Unit: "SET", SourceItemNo: "10", Pages: []int{2}}, {Name: "Labor", Quantity: "2", Unit: "SET", SourceItemNo: "20", Pages: []int{2}}}, nil
}

type lostAutomatic struct {
	Objects feature.Objects
	lost    atomic.Bool
}

func (o *lostAutomatic) Create(ctx context.Context, a, r, k string, c objectDomainEntity.Content) (objectDomainEntity.Object, error) {
	created, e := o.Objects.Create(ctx, a, r, k, c)
	if e == nil && o.lost.CompareAndSwap(false, true) {
		return objectDomainEntity.Object{}, errors.New("private provider sentinel")
	}
	return created, e
}

func TestEmulatorAutomaticLostReplyAndPartialResume(t *testing.T) {
	ctx, client, r, parent := importEmulator(t)
	manual, _, objects := importServices(client, r)
	p := &autoProducer{}
	lost := &lostAutomatic{Objects: objects}
	s := &automatic.Service{Repositories: manual.Repositories, Objects: lost, Store: &AutomaticStore{Store: Store{Client: client, ChannelID: "isolated-import-tests"}}, Producer: p}
	source := feature.Source{SourceDocument: "pdf:" + strings.Repeat("a", 64), SourceRevision: "", Pages: []feature.Page{{Number: 2, Text: "whole-page-private-sentinel item10 pump; item20 labor"}}}
	key := "automatic-key-0001"
	first, e := s.Create(ctx, "actor-a", r.ID, key, source, "work")
	if e != nil || first.Status != "processing" || first.Reason != "creation_unavailable" || importCount(t, ctx, parent, "objects") != 1 {
		t.Fatal(first, e)
	}
	// The caller can lose the first response entirely: key + full source recovers it.
	got, e := s.Create(ctx, "actor-a", r.ID, key, source, "work")
	if e != nil || got.Status != "completed" || p.calls.Load() != 1 || importCount(t, ctx, parent, "objects") != 2 {
		t.Fatal(got, e, p.calls.Load())
	}
	replay, e := s.Create(ctx, "actor-a", r.ID, key, source, "work")
	if e != nil || replay.ID != got.ID || p.calls.Load() != 1 {
		t.Fatal(replay, e)
	}
	read, e := s.Get(ctx, "actor-a", r.ID, got.ID)
	if e != nil || read.Status != "completed" {
		t.Fatal(read, e)
	}
	changed := source
	changed.SourceRevision = "changed"
	if _, e = s.Create(ctx, "actor-a", r.ID, key, changed, "work"); !errors.Is(e, feature.ErrCreationConflict) {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, "actor-b", r.ID, got.ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal("other creator", e)
	}
	if _, e = s.Get(ctx, "outsider", r.ID, got.ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal("nonmember", e)
	}
	if _, e = s.Get(ctx, "actor-a", strings.Repeat("c", 32), got.ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal("cross repository", e)
	}
	snap, e := parent.Collection("objectImportRuns").Doc(got.ID).Get(ctx)
	if e != nil {
		t.Fatal(e)
	}
	var d document
	if e = snap.DataTo(&d); e != nil || strings.Contains(d.Payload, "whole-page-private-sentinel") {
		t.Fatal("retained source", e)
	}
	second, e := s.Create(ctx, "actor-a", r.ID, "automatic-key-0002", source, "work")
	if e != nil || second.ID == got.ID || importCount(t, ctx, parent, "objects") != 4 {
		t.Fatal("different keys", e)
	}
	if _, e = parent.Update(ctx, []db.Update{{Path: "status", Value: "ARCHIVED"}}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, "actor-a", r.ID, got.ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatal("inactive repository", e)
	}
}

func TestEmulatorAutomaticConcurrentFirstFreezeWins(t *testing.T) {
	ctx, client, r, parent := importEmulator(t)
	manual, _, objects := importServices(client, r)
	barrier := &sync.WaitGroup{}
	barrier.Add(2)
	p := &autoProducer{barrier: barrier}
	s := &automatic.Service{Repositories: manual.Repositories, Objects: objects, Store: &AutomaticStore{Store: Store{Client: client, ChannelID: "isolated-import-tests"}}, Producer: p}
	source := feature.Source{SourceDocument: "synthetic", Pages: []feature.Page{{Number: 2, Text: "pump labor"}}}
	results := make([]automatic.Run, 2)
	errs := make([]error, 2)
	var group sync.WaitGroup
	for i := range 2 {
		group.Go(func() { results[i], errs[i] = s.Create(ctx, "actor-a", r.ID, "concurrent-auto-key", source, "work") })
	}
	group.Wait()
	for i, e := range errs {
		if e != nil || results[i].Status != "completed" {
			t.Fatal(results[i], e)
		}
	}
	if p.calls.Load() != 2 || results[0].Rows[0].Name != results[1].Rows[0].Name || results[0].Rows[0].ResultObjectID != results[1].Rows[0].ResultObjectID || importCount(t, ctx, parent, "objects") != 2 {
		t.Fatal("losing model output used", results)
	}
}

func TestAutomaticStorageReservesEveryResultAndRoundtrips(t *testing.T) {
	now := time.Now().UTC()
	j := automatic.Run{ID: strings.Repeat("a", 32), RepositoryID: strings.Repeat("b", 32), CreatorID: "actor", IntentHash: strings.Repeat("c", 64), SourceDocument: "synthetic", SourceHash: strings.Repeat("d", 64), Status: "processing", CreatedAt: now, UpdatedAt: now}
	for i := range 100 {
		j.Rows = append(j.Rows, automatic.Row{ID: fmt.Sprintf("%032x", i+1), Fields: feature.Fields{Name: "Pump", Description: strings.Repeat("x", 2000), Type: "Equipment", Quantity: "1", Unit: "SET"}, SourceItemNo: fmt.Sprint(i + 1), SourcePages: "7"})
	}
	// Binary search the stored byte threshold, with room for every final result.
	low, high := 0, 8000
	for low < high {
		mid := (low + high + 1) / 2
		for i := range j.Rows {
			j.Rows[i].Description = strings.Repeat("x", mid)
		}
		if _, e := encodeAutomatic(j); e == nil {
			low = mid
		} else {
			high = mid - 1
		}
	}
	for i := range j.Rows {
		j.Rows[i].Description = strings.Repeat("x", low)
	}
	for {
		if _, e := encodeAutomatic(j); e != nil {
			break
		}
		j.Rows[0].Description += "x"
	}
	j.Rows[0].Description = j.Rows[0].Description[:len(j.Rows[0].Description)-1]
	doc, e := encodeAutomatic(j)
	if e != nil {
		t.Fatal(e)
	}
	var stored storedAutomaticRun
	if json.Unmarshal([]byte(doc.Payload), &stored) != nil {
		t.Fatal("DTO decode")
	}
	got := fromStoredRun(stored)
	if !reflect.DeepEqual(j, got) {
		t.Fatal("DTO drift")
	}
	for i := range j.Rows {
		j.Rows[i].ResultObjectID = strings.Repeat("f", 32)
	}
	j.Status = "completed"
	if _, e = encodeAutomatic(j); e != nil {
		t.Fatal("accepted snapshot lacks final result budget", e)
	}
}

type workProducer struct{ calls atomic.Int32 }

func (p *workProducer) Rows(context.Context, feature.Source) ([]automatic.Candidate, error) {
	p.calls.Add(1)
	return []automatic.Candidate{{Name: "Install", Quantity: "11", Unit: "SET", SourceItemNo: "10", Pages: []int{2}}}, nil
}

func TestEmulatorAutomaticWorkInitializationAndRetry(t *testing.T) {
	ctx, client, r, _ := importEmulator(t)
	manual, _, objects := importServices(client, r)
	p := &workProducer{}
	s := &automatic.Service{Repositories: manual.Repositories, Objects: objects, Store: &AutomaticStore{Store: Store{Client: client, ChannelID: "isolated-import-tests"}}, Producer: p}
	source := feature.Source{SourceDocument: "synthetic", Pages: []feature.Page{{Number: 2, Text: "install eleven"}}}
	run, e := s.Create(ctx, "actor-a", r.ID, "automatic-work-key", source, "work")
	if e != nil || run.Status != "completed" {
		t.Fatal(run, e)
	}
	created, e := objects.Get(ctx, "actor-a", r.ID, run.Rows[0].ResultObjectID)
	if e != nil || created.CompletedQuantity != "0" || created.Status != objectDomainEntity.StatusNotStarted {
		t.Fatal(created, e)
	}
	retry, e := s.Create(ctx, "actor-a", r.ID, "automatic-work-key", source, "work")
	if e != nil || retry.Rows[0].ResultObjectID != created.ID || p.calls.Load() != 1 {
		t.Fatal(retry, e)
	}
}

func TestEmulatorSelectedTypeAndLegacyIntentCompatibility(t *testing.T) {
	ctx, client, r, parent := importEmulator(t)
	manual, _, objects := importServices(client, r)
	p := &autoProducer{}
	store := &AutomaticStore{Store: Store{Client: client, ChannelID: "isolated-import-tests"}}
	s := &automatic.Service{Repositories: manual.Repositories, Objects: objects, Store: store, Producer: p}
	source := feature.Source{SourceDocument: "synthetic", SourceRevision: "v0", Pages: []feature.Page{{Number: 2, Text: "synthetic eleven"}}}
	if _, e := s.Create(ctx, "actor-a", r.ID, "missing-batch-type", source, ""); !errors.Is(e, automatic.ErrTypeRequired) || p.calls.Load() != 0 || importCount(t, ctx, parent, "objectImportRuns") != 0 {
		t.Fatal("new missing selection wrote or parsed", e)
	}
	for _, typ := range []string{objectDomainEntity.TypeWork, objectDomainEntity.TypeNonWork} {
		key := "batch-type-key-" + typ
		run, e := s.Create(ctx, "actor-a", r.ID, key, source, typ)
		if e != nil || run.Status != "completed" {
			t.Fatal(run, e)
		}
		for _, row := range run.Rows {
			created, e := objects.Get(ctx, "actor-a", r.ID, row.ResultObjectID)
			if e != nil || created.Type != typ || created.Description != "" || created.MaterialCode != "" {
				t.Fatal(created, e)
			}
			if typ == objectDomainEntity.TypeWork && created.CompletedQuantity != "0" || typ == objectDomainEntity.TypeNonWork && created.CompletedQuantity != "" || created.Status != objectDomainEntity.StatusNotStarted {
				t.Fatal("wrong progress", created)
			}
		}
		other := objectDomainEntity.TypeWork
		if typ == other {
			other = objectDomainEntity.TypeNonWork
		}
		if _, e = s.Create(ctx, "actor-a", r.ID, key, source, other); !errors.Is(e, feature.ErrCreationConflict) {
			t.Fatal("different selected type accepted", e)
		}
	}
	_, digest, e := feature.ValidateSource(source)
	if e != nil {
		t.Fatal(e)
	}
	seed := func(key, status string) automatic.Run {
		t.Helper()
		now := time.Now().UTC().Truncate(time.Microsecond)
		j := automatic.Run{ID: feature.Hash([]string{"automatic-import-v1", "actor-a", r.ID, key})[:32], RepositoryID: r.ID, CreatorID: "actor-a", IntentHash: feature.Hash([]string{"automatic-intent-v1", source.SourceDocument, source.SourceRevision, digest}), SourceDocument: source.SourceDocument, SourceRevision: source.SourceRevision, SourceHash: digest, Status: status, Rows: []automatic.Row{}, CreatedAt: now, UpdatedAt: now}
		if status == "processing" {
			j.Rows = []automatic.Row{{Fields: feature.Fields{Name: "Historical", Description: "old description", MaterialCode: "old material", Type: "Engineering", Quantity: "11", Unit: "SET"}, ID: strings.Repeat("f", 32), SourceItemNo: "10", SourcePages: "2"}}
		}
		saved, e := store.Update(ctx, r, "actor-a", j.ID, func(*automatic.Run) (automatic.Run, error) { return j, nil })
		if e != nil {
			t.Fatal(e)
		}
		return saved
	}
	old := seed("legacy-frozen-key", "processing")
	calls := p.calls.Load()
	resumed, e := s.Create(ctx, "actor-a", r.ID, "legacy-frozen-key", source, "")
	if e != nil || resumed.Status != "completed" || p.calls.Load() != calls || resumed.Rows[0].Type != old.Rows[0].Type {
		t.Fatal("old frozen not resumed", resumed, e)
	}
	created, e := objects.Get(ctx, "actor-a", r.ID, resumed.Rows[0].ResultObjectID)
	if e != nil || created.Description != "old description" || created.MaterialCode != "old material" || created.Type != "Engineering" {
		t.Fatal("historical fields lost", created, e)
	}
	if _, e = s.Create(ctx, "actor-a", r.ID, "legacy-frozen-key", source, objectDomainEntity.TypeWork); !errors.Is(e, feature.ErrCreationConflict) {
		t.Fatal("old frozen retyped", e)
	}
	pending := seed("legacy-pending-key", "pending")
	ref := parent.Collection("objectImportRuns").Doc(pending.ID)
	before, e := ref.Get(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Create(ctx, "actor-a", r.ID, "legacy-pending-key", source, ""); !errors.Is(e, automatic.ErrTypeRequired) {
		t.Fatal("legacy pending parsed", e)
	}
	after, e := ref.Get(ctx)
	if e != nil || !before.UpdateTime.Equal(after.UpdateTime) || p.calls.Load() != calls {
		t.Fatal("pending legacy mutated", e)
	}
	if _, e = s.Create(ctx, "actor-a", r.ID, "legacy-pending-key", source, objectDomainEntity.TypeWork); !errors.Is(e, feature.ErrCreationConflict) {
		t.Fatal("legacy pending silently upgraded", e)
	}
}
