package firestore

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type templateRepositories struct {
	r      repository.Repository
	denied bool
}

func (p *templateRepositories) Get(context.Context, string, string) (repository.Repository, error) {
	if p.denied {
		return repository.Repository{}, repository.ErrNotFound
	}
	return p.r, nil
}

func TestEmulatorTemplateClosedLoop(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	repos := &templateRepositories{r: r}
	service := application.Service{Repositories: repos, Store: store, Templates: store}
	input := dto.TemplateCreation{Name: "工程範本", Nodes: []template.Node{{ID: "leaf", ParentID: "child", Name: "子工序", Type: "work", Quantity: "2.0", Unit: "組"}, {ID: "root", Name: "工程"}, {ID: "child", ParentID: "root", Name: "工序"}}}
	snapshot, e := service.CreateTemplate(ctx, "actor", r.ID, "create-template-key", input)
	if e != nil {
		t.Fatal(e)
	}
	replay, e := service.CreateTemplate(ctx, "actor", r.ID, "create-template-key", input)
	if e != nil || !reflect.DeepEqual(snapshot, replay) {
		t.Fatal(replay, e)
	}
	input.Name = "different"
	if _, e = service.CreateTemplate(ctx, "actor", r.ID, "create-template-key", input); !errors.Is(e, entity.ErrCreationConflict) {
		t.Fatal(e)
	}
	page, e := service.ListTemplates(ctx, "actor", r.ID, "")
	if e != nil || len(page.Items) != 1 {
		t.Fatal(page, e)
	}
	read, e := service.GetTemplate(ctx, "actor", r.ID, snapshot.ID)
	if e != nil || !reflect.DeepEqual(read, snapshot) {
		t.Fatal(read, e)
	}
	const n = 5
	results := make([]dto.TemplateInstance, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			results[i], errs[i] = service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "apply-template-key", "")
		})
	}
	wg.Wait()
	for i := range n {
		if errs[i] != nil || !reflect.DeepEqual(results[0], results[i]) {
			t.Fatalf("replay %d: %+v %v", i, results[i], errs[i])
		}
	}
	first := results[0]
	if len(first.ObjectIDs) != 3 || len(first.Nodes) != 3 || collectionCount(t, ctx, store.repositoryRef(r.ID), "objects") != 3 {
		t.Fatal(first)
	}
	for i, id := range first.ObjectIDs {
		obj, e := service.Get(ctx, "actor", r.ID, id)
		if e != nil {
			t.Fatal(e)
		}
		if obj.TemplateID != snapshot.ID || obj.TemplateVersion != 1 || obj.TemplateNodeID != snapshot.Nodes[i].ID {
			t.Fatal(obj)
		}
		if i == 0 && (obj.ParentID != first.ObjectIDs[2] || obj.CompletedQuantity != "0" || obj.Quantity != "2") {
			t.Fatal(obj)
		}
	}
	second, e := service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "another-apply-key", first.RootObjectID)
	if e != nil || second.RootObjectID == first.RootObjectID {
		t.Fatal(second, e)
	}
	root, e := service.Get(ctx, "actor", r.ID, second.RootObjectID)
	if e != nil || root.ParentID != first.RootObjectID {
		t.Fatal(root, e)
	}
	_, e = service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "apply-template-key", first.RootObjectID)
	if !errors.Is(e, entity.ErrCreationConflict) {
		t.Fatal(e)
	}
	before := collectionCount(t, ctx, store.repositoryRef(r.ID), "objects")
	_, e = service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "missing-parent-key", strings.Repeat("f", 32))
	if !errors.Is(e, entity.ErrNotFound) {
		t.Fatal(e)
	}
	if collectionCount(t, ctx, store.repositoryRef(r.ID), "objects") != before {
		t.Fatal("partial tree")
	}
	repos.denied = true
	_, e = service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "denied-apply-key", "")
	if !errors.Is(e, entity.ErrNotFound) {
		t.Fatal(e)
	}
	repos.denied = false
	// Ordinary updates preserve provenance and do not mutate the template or other instances.
	obj, e := service.Get(ctx, "actor", r.ID, first.ObjectIDs[0])
	if e != nil {
		t.Fatal(e)
	}
	content := obj.Content
	content.Name = "獨立修改"
	content.CompletedQuantity = "1"
	one := "1"
	updated, e := service.Update(ctx, "actor", r.ID, obj.ID, entity.Update{Content: content, CompletedQuantity: &one, ExpectedVersion: 1})
	if e != nil || updated.TemplateID != snapshot.ID || updated.TemplateNodeID != "leaf" {
		t.Fatal(updated, e)
	}
	unchanged, e := service.Get(ctx, "actor", r.ID, second.ObjectIDs[0])
	if e != nil || unchanged.CompletedQuantity != "0" {
		t.Fatal(unchanged, e)
	}
	final, e := service.GetTemplate(ctx, "actor", r.ID, snapshot.ID)
	if e != nil || !reflect.DeepEqual(final, snapshot) {
		t.Fatal(final, e)
	}
	lostResponse, e := service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "apply-template-key", "")
	if e != nil || !reflect.DeepEqual(lostResponse, first) {
		t.Fatal(lostResponse, e)
	}
}

func TestEmulatorTemplateRollbackAfterBufferedWrites(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	repos := &templateRepositories{r: r}
	service := application.Service{Repositories: repos, Store: store, Templates: store}
	snapshot, e := service.CreateTemplate(ctx, "actor", r.ID, "rollback-create-key", dto.TemplateCreation{Name: "tree", Nodes: []template.Node{{ID: "root", Name: "root"}, {ID: "child", ParentID: "root", Name: "child"}}})
	if e != nil {
		t.Fatal(e)
	}
	plan := []entity.Content{{Name: "root"}, {Name: ""}}
	// The second node fails after the first tx.Create has buffered a write.
	_, e = store.InstantiateTemplate(ctx, r, "actor", "rollback-apply-key", snapshot, "", plan)
	if !errors.Is(e, entity.ErrInput) {
		t.Fatal(e)
	}
	for _, name := range []string{"objects", "objectTemplateInstanceReceipts"} {
		if collectionCount(t, ctx, store.repositoryRef(r.ID), name) != 0 {
			t.Fatalf("partial %s after aborted transaction", name)
		}
	}
	result, e := service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "rollback-apply-key", "")
	if e != nil || len(result.ObjectIDs) != 2 {
		t.Fatal(result, e)
	}
}

func TestEmulatorEditedTemplate(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	service := application.Service{Repositories: &templateRepositories{r: r}, Store: store, Templates: store}
	source := []template.Node{
		{ID: "root", Name: "電力系統"},
		{ID: "cable", ParentID: "root", Name: "高壓線"},
		{ID: "branch", ParentID: "root", Name: "接地工程"},
		{ID: "leaf", ParentID: "branch", Name: "接地測試"},
	}
	snapshot, err := service.CreateTemplate(ctx, "actor", r.ID, "create-edited-template", dto.TemplateCreation{Name: "電力系統", Nodes: source})
	if err != nil {
		t.Fatal(err)
	}
	apply := func(nodes []template.Node) (dto.TemplateInstance, error) {
		return service.InstantiateTemplate(ctx, "actor", r.ID, snapshot.ID, "apply-edited-template", "", nodes)
	}
	if _, err := apply([]template.Node{source[0], source[3]}); !errors.Is(err, entity.ErrInput) {
		t.Fatalf("orphan should be rejected: %v", err)
	}
	if got := collectionCount(t, ctx, store.repositoryRef(r.ID), "objects"); got != 0 {
		t.Fatalf("invalid request wrote %d objects", got)
	}
	edits := append([]template.Node(nil), source[:2]...)
	edits[1].Name, edits[1].Quantity, edits[1].Unit = "高壓線敷設", "6400.0", "米"
	first, err := apply(edits)
	if err != nil || len(first.ObjectIDs) != 2 {
		t.Fatalf("apply: %+v %v", first, err)
	}
	got, err := service.Get(ctx, "actor", r.ID, first.ObjectIDs[1])
	if err != nil || got.Name != "高壓線敷設" || got.Quantity != "6400" || got.Unit != "米" || got.ParentID != first.RootObjectID {
		t.Fatalf("read back: %+v %v", got, err)
	}
	replay, err := apply(edits)
	if err != nil || !reflect.DeepEqual(first, replay) {
		t.Fatalf("replay: %+v %v", replay, err)
	}
	edits[1].Quantity = "6500"
	if _, err := apply(edits); !errors.Is(err, entity.ErrCreationConflict) {
		t.Fatalf("changed request should conflict: %v", err)
	}
	if got := collectionCount(t, ctx, store.repositoryRef(r.ID), "objects"); got != 2 {
		t.Fatalf("want exactly two objects, got %d", got)
	}
	unchanged, err := service.GetTemplate(ctx, "actor", r.ID, snapshot.ID)
	if err != nil || !reflect.DeepEqual(snapshot, unchanged) {
		t.Fatalf("source template changed: %+v %v", unchanged, err)
	}
}

func TestEmulatorPersonalTemplateAcrossRepositories(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	repos := &templateRepositories{r: r}
	service := application.Service{Repositories: repos, Store: store, Templates: store, PersonalTemplates: store}
	nodes := []template.Node{{ID: "root", Name: "電力系統"}, {ID: "cable", ParentID: "root", Name: "線材", Unit: "米"}}
	created, err := service.CreatePersonalTemplate(ctx, "alice", "create-personal-key", dto.TemplateCreation{Name: "電力", Nodes: nodes})
	if err != nil {
		t.Fatal(err)
	}
	for actor, want := range map[string]int{"alice": 1, "bob": 0} {
		page, err := service.ListPersonalTemplates(ctx, actor, "")
		if err != nil || len(page.Items) != want {
			t.Fatalf("list %s: %+v %v", actor, page, err)
		}
	}
	apply := func(actor string, target repository.Repository, plan []template.Node) (dto.TemplateInstance, error) {
		repos.r = target
		return service.InstantiatePersonalTemplate(ctx, actor, target.ID, created.ID, "apply-personal-key", "", plan)
	}
	if _, err := apply("bob", r, nodes); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("foreign source: %v", err)
	}
	repos.denied = true
	if _, err := apply("alice", r, nodes); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("foreign target: %v", err)
	}
	repos.denied = false
	if got := collectionCount(t, ctx, store.repositoryRef(r.ID), "objects"); got != 0 {
		t.Fatalf("denied writes: %d", got)
	}
	first, err := apply("alice", r, nodes[:1])
	if err != nil || len(first.ObjectIDs) != 1 {
		t.Fatalf("apply first: %+v %v", first, err)
	}
	r2 := r
	r2.ID = strings.Repeat("b", 32)
	if _, err := store.repositoryRef(r2.ID).Set(ctx, map[string]any{"id": r2.ID, "ownerId": r2.OwnerID, "status": r2.Status}); err != nil {
		t.Fatal(err)
	}
	nodes[1].Quantity = "6400"
	second, err := apply("alice", r2, nodes)
	if err != nil || len(second.ObjectIDs) != 2 || second.RootObjectID == first.RootObjectID {
		t.Fatalf("apply second: %+v %v", second, err)
	}
	got, err := service.Get(ctx, "alice", r2.ID, second.ObjectIDs[1])
	if err != nil || got.Quantity != "6400" || got.Unit != "米" || got.ParentID != second.RootObjectID {
		t.Fatalf("read back: %+v %v", got, err)
	}
	replay, err := apply("alice", r2, nodes)
	if err != nil || !reflect.DeepEqual(second, replay) {
		t.Fatalf("replay: %+v %v", replay, err)
	}
	nodes[1].Quantity = "6500"
	if _, err := apply("alice", r2, nodes); !errors.Is(err, entity.ErrCreationConflict) {
		t.Fatalf("changed replay: %v", err)
	}
	if got := collectionCount(t, ctx, store.repositoryRef(r2.ID), "objects"); got != 2 {
		t.Fatalf("expected two objects: %d", got)
	}
	source, err := store.GetPersonalTemplate(ctx, "alice", created.ID)
	if err != nil || !reflect.DeepEqual(created.Nodes, source.Nodes) {
		t.Fatalf("source changed: %+v %v", source, err)
	}
}

func TestEmulatorPersonalTemplateRevision(t *testing.T) {
	ctx, store, r := emulatorStore(t)
	s := application.Service{Repositories: &templateRepositories{r: r}, PersonalTemplates: store, Templates: store, Store: store}
	first := dto.TemplateCreation{Name: "first", Nodes: []template.Node{{ID: "root", Name: "root"}}}
	original, err := s.CreatePersonalTemplate(ctx, "alice", "original-request", first)
	if err != nil {
		t.Fatal(err)
	}
	old, err := s.InstantiatePersonalTemplate(ctx, "alice", r.ID, original.ID, "apply-original-key", "", original.Nodes)
	if err != nil {
		t.Fatal(err)
	}
	edit := first
	edit.TemplateID = original.ID
	edit.BaseVersion = 1
	edit.Name = "second"
	updated, err := s.CreatePersonalTemplate(ctx, "alice", "revision-request", edit)
	if err != nil || updated.ID != original.ID || updated.Version != 2 {
		t.Fatalf("revision %+v %v", updated, err)
	}
	if _, err = s.CreatePersonalTemplate(ctx, "alice", "stale-request-key", edit); !errors.Is(err, entity.ErrConflict) {
		t.Fatalf("stale edit: %v", err)
	}
	replay, err := s.CreatePersonalTemplate(ctx, "alice", "original-request", first)
	if err != nil || !reflect.DeepEqual(replay, original) {
		t.Fatalf("original replay %+v %v", replay, err)
	}
	replay, err = s.CreatePersonalTemplate(ctx, "alice", "revision-request", edit)
	if err != nil || !reflect.DeepEqual(replay, updated) {
		t.Fatalf("revision replay %+v %v", replay, err)
	}
	page, err := s.ListPersonalTemplates(ctx, "alice", "")
	if err != nil || len(page.Items) != 1 || page.Items[0].Version != 2 {
		t.Fatalf("list %+v %v", page, err)
	}
	obj, err := s.Get(ctx, "alice", r.ID, old.RootObjectID)
	if err != nil || obj.TemplateVersion != 1 {
		t.Fatalf("old object %+v %v", obj, err)
	}
	newer, err := s.InstantiatePersonalTemplate(ctx, "alice", r.ID, original.ID, "apply-new-version", "", updated.Nodes, 2)
	if err != nil || newer.TemplateVersion != 2 {
		t.Fatalf("new instance %+v %v", newer, err)
	}
}
