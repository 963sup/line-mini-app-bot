package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/object/application/command"
	"github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/object/domain/template"
	repository "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type templateStoreStub struct {
	calls    int
	snapshot dto.Template
	plan     []entity.Content
}

func (s *templateStoreStub) CreateTemplate(_ context.Context, _ repository.Repository, _, _ string, in dto.TemplateCreation) (dto.Template, error) {
	s.calls++
	return dto.Template{Name: in.Name, Nodes: in.Nodes}, nil
}

func (s *templateStoreStub) ListTemplates(context.Context, repository.Repository, string, string) (dto.TemplatePage, error) {
	s.calls++
	return dto.TemplatePage{}, nil
}

func (s *templateStoreStub) GetTemplate(context.Context, repository.Repository, string) (dto.Template, error) {
	s.calls++
	return s.snapshot, nil
}

func (s *templateStoreStub) InstantiateTemplate(_ context.Context, _ repository.Repository, _, _ string, _ dto.Template, _ string, plan []entity.Content, _ ...[]template.Node) (dto.TemplateInstance, error) {
	s.calls++
	s.plan = plan
	return dto.TemplateInstance{}, nil
}

func TestTemplateValidationAndAuthority(t *testing.T) {
	rid := strings.Repeat("a", 32)
	base := dto.TemplateCreation{Name: " 範本 ", Nodes: []template.Node{{ID: "root", Name: " 根 "}, {ID: "child", ParentID: "root", Name: " 施工 ", Type: "work", Quantity: "2.000000", Unit: "組"}}}
	normalized, e := command.NormalizeTemplate(base)
	if e != nil || normalized.Name != "範本" || normalized.Nodes[1].Quantity != "2" {
		t.Fatal(normalized, e)
	}
	if base.Nodes[1].Name != " 施工 " {
		t.Fatal("mutated input")
	}
	cases := []dto.TemplateCreation{
		{Name: "", Nodes: base.Nodes},
		{Name: "x"},
		{Name: "x", Nodes: []template.Node{{ID: "r", Name: "x"}, {ID: "s", Name: "y"}}},
		{Name: "x", Nodes: []template.Node{{ID: "r", Name: "x"}, {ID: "s", ParentID: "s", Name: "y"}}},
		{Name: "x", Nodes: []template.Node{{ID: "r", Name: "x", Type: "work", Quantity: "2"}}},
		{Name: "x", Nodes: []template.Node{{ID: "r", Name: "x", Quantity: "-1"}}},
		{Name: "x", Nodes: []template.Node{{ID: "r", Name: "x"}, {ID: "s", ParentID: "missing", Name: "y"}}},
	}
	for _, in := range cases {
		if _, e := command.NormalizeTemplate(in); !errors.Is(e, entity.ErrInput) {
			t.Fatalf("accepted %+v: %v", in, e)
		}
	}
	tooLarge := dto.TemplateCreation{Name: "x"}
	for i := range 101 {
		tooLarge.Nodes = append(tooLarge.Nodes, template.Node{ID: strings.Repeat("a", i+1), Name: "x"})
	}
	if _, e := command.NormalizeTemplate(tooLarge); !errors.Is(e, entity.ErrInput) {
		t.Fatal(e)
	}
	store := &templateStoreStub{snapshot: dto.Template{ID: rid, RepositoryID: rid, Version: 1, Nodes: normalized.Nodes}}
	repo := repository.Repository{ID: rid, OwnerType: repository.OwnerUser, OwnerID: "U" + rid, Status: "ACTIVE"}
	service := Service{Repositories: repositories{r: repo}, Store: &stubStore{}, Templates: store}
	if _, e := service.InstantiateTemplate(t.Context(), "actor", rid, rid, strings.Repeat("k", 16), ""); e != nil {
		t.Fatal(e)
	}
	if store.plan[1].CompletedQuantity != "0" || store.plan[1].Quantity != "2" {
		t.Fatal(store.plan)
	}
	service.Repositories = repositories{e: repository.ErrNotFound}
	before := store.calls
	_, e1 := service.CreateTemplate(t.Context(), "actor", rid, strings.Repeat("k", 16), base)
	_, e2 := service.ListTemplates(t.Context(), "actor", rid, "")
	_, e3 := service.GetTemplate(t.Context(), "actor", rid, rid)
	_, e4 := service.InstantiateTemplate(t.Context(), "actor", rid, rid, strings.Repeat("k", 16), "")
	for _, e := range []error{e1, e2, e3, e4} {
		if !errors.Is(e, entity.ErrNotFound) {
			t.Fatal(e)
		}
	}
	if store.calls != before {
		t.Fatal("unauthorized store call")
	}
}
