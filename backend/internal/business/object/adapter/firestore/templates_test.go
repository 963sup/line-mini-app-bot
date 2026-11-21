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
