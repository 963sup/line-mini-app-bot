package project_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	p "github.com/963sup/line-mini-app-bot/internal/business/project"
	sp "github.com/963sup/line-mini-app-bot/internal/business/space"
)

type spaces struct{ denied bool }

func (s *spaces) Get(_ context.Context, a, id string) (sp.Space, error) {
	if s.denied || a == "outsider" {
		return sp.Space{}, sp.ErrNotFound
	}
	return sp.Space{ID: id, GroupID: id, Status: "ACTIVE"}, nil
}

func (s *spaces) LegacyGet(ctx context.Context, a, id string) (sp.Space, error) {
	return s.Get(ctx, a, id)
}

type sources struct {
	denied bool
	space  string
}

func (s *sources) Authorize(_ context.Context, a, id, owner string) (p.Source, error) {
	if s.denied || a == "outsider" || owner != "" && owner != s.space {
		return p.Source{}, p.ErrNotFound
	}
	return p.Source{ID: id, SpaceID: s.space}, nil
}

type store struct {
	item     p.Project
	progress int
}

func (s *store) List(context.Context, p.Scope, string, string) (p.Page, error) {
	return p.Page{Items: []p.Project{s.item}}, nil
}
func (s *store) Get(context.Context, p.Scope, string) (p.Project, error) { return s.item, nil }
func (s *store) Create(_ context.Context, r p.Scope, a, k string, c p.Content) (p.Project, error) {
	s.item = p.Project{ID: strings.Repeat("2", 32), SpaceID: r.SpaceID, SourceRepositoryID: c.SourceRepositoryID, CreatorID: a, Title: c.Title, Body: c.Body, Version: 1}
	return s.item, nil
}

func (s *store) Update(_ context.Context, r p.Scope, a, id string, in p.Update) (p.Project, error) {
	v, e := p.ApplyUpdate(s.item, a, in)
	if e == nil {
		s.item = v
	}
	return v, e
}

func (s *store) ListItems(context.Context, p.Scope, string, string, string) (p.ItemPage, error) {
	return p.ItemPage{}, nil
}

func (s *store) ChangeItem(context.Context, p.Scope, string, string, string, bool, int64) (p.Project, error) {
	return s.item, nil
}
func (s *store) Progress(context.Context, p.Scope, *p.Project) error { s.progress++; return nil }
func TestSpaceOwnedMetadataIndependentSource(t *testing.T) {
	ctx := context.Background()
	gid := "C" + strings.Repeat("1", 32)
	rid := strings.Repeat("3", 32)
	pid := strings.Repeat("2", 32)
	ss := &spaces{}
	src := &sources{space: gid}
	st := &store{item: p.Project{ID: pid, SpaceID: gid, SourceRepositoryID: rid, CreatorID: "owner", Title: "own metadata", Version: 1}}
	svc := &p.Service{Spaces: ss, Sources: src, Store: st}
	if _, err := svc.Create(ctx, "owner", "C"+strings.Repeat("5", 32), "cross-space-key-0001", p.Content{Title: "must reject foreign source", SourceRepositoryID: rid}); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("Space membership must not authorize another Space source: %v", err)
	}
	src.denied = true
	got, e := svc.Get(ctx, "owner", gid, pid)
	if e != nil || got.Title != "own metadata" || got.SourceAvailable || st.progress != 0 {
		t.Fatalf("own metadata must survive source denial without progress: %+v %v calls=%d", got, e, st.progress)
	}
	page, e := svc.List(ctx, "owner", gid, "")
	if e != nil || len(page.Items) != 1 || page.Items[0].SourceAvailable {
		t.Fatalf("Space list must remain own metadata: %+v %v", page, e)
	}
	if _, e = svc.ListItems(ctx, "owner", gid, pid, ""); !errors.Is(e, p.ErrNotFound) {
		t.Fatalf("source item read must deny after source revoke: %v", e)
	}
	if _, e = svc.ChangeItem(ctx, "owner", gid, pid, strings.Repeat("4", 32), true, 1); !errors.Is(e, p.ErrNotFound) {
		t.Fatalf("source item mutation must deny after revoke: %v", e)
	}
	_, e = svc.Update(ctx, "owner", gid, pid, p.Update{Content: p.Content{Title: "edited own metadata"}, ExpectedVersion: 1})
	if e != nil {
		t.Fatalf("own metadata update must not require source: %v", e)
	}
	if _, e = svc.Update(ctx, "member", gid, pid, p.Update{Content: p.Content{Title: "forbidden"}, ExpectedVersion: 2}); !errors.Is(e, p.ErrForbidden) {
		t.Fatalf("member cannot manage another Project: %v", e)
	}
	ss.denied = true
	src.denied = false
	if _, e = svc.List(ctx, "owner", gid, ""); !errors.Is(e, p.ErrNotFound) {
		t.Fatalf("canonical must deny Space revoke: %v", e)
	}
	if _, e = svc.List(ctx, "owner", rid, ""); !errors.Is(e, p.ErrNotFound) {
		t.Fatalf("legacy cannot bypass Space revoke using source auth: %v", e)
	}
}
