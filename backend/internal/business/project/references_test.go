package project_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	p "github.com/963sup/line-mini-app-bot/internal/business/project"
)

type referenceStore struct {
	store
	hydrate bool
	changes int
}

func (s *referenceStore) ListReferences(_ context.Context, _ p.Scope, _, _, _ string, hydrate bool) (p.ReferencePage, error) {
	s.hydrate = hydrate
	return p.ReferencePage{Items: []p.Reference{{SourceType: "object", SourceUnavailable: !hydrate}}}, nil
}

func (s *referenceStore) ChangeReference(context.Context, p.Scope, string, string, p.Reference, bool, int64) (p.Project, error) {
	s.changes++
	return s.item, nil
}

func TestReferenceIdentityAndIndependentAuthorization(t *testing.T) {
	id := strings.Repeat("a", 32)
	for _, kind := range []string{"object", "issue", "discussion"} {
		if !p.ValidReference(p.Reference{SourceType: kind, SourceID: id}) || p.ValidReference(p.Reference{SourceType: kind, SourceID: id, ParentObjectID: id}) {
			t.Fatal(kind)
		}
	}
	if p.ValidReference(p.Reference{SourceType: "subobject", SourceID: id}) || !p.ValidReference(p.Reference{SourceType: "subobject", SourceID: id, ParentObjectID: id}) {
		t.Fatal("child binding")
	}
	gid := "C" + id
	st := &referenceStore{store: store{item: p.Project{ID: id, SpaceID: gid, SourceRepositoryID: id, CreatorID: "owner", Version: 1}}}
	src := &sources{denied: true, space: gid}
	svc := &p.Service{Spaces: &spaces{}, Sources: src, Store: st}
	page, e := svc.ListReferences(context.Background(), "owner", gid, id, "")
	if e != nil || st.hydrate || !page.Items[0].SourceUnavailable {
		t.Fatal(page, e)
	}
	if _, e = svc.AddReference(context.Background(), "owner", gid, id, p.Reference{SourceType: "object", SourceID: id}, 1); !errors.Is(e, p.ErrNotFound) || st.changes != 0 {
		t.Fatal("source denial bypass", e)
	}
	if _, e = svc.RemoveReference(context.Background(), "member", gid, id, id, 1); !errors.Is(e, p.ErrForbidden) || st.changes != 0 {
		t.Fatal("owner denial bypass", e)
	}
	if _, e = svc.RemoveReference(context.Background(), "owner", gid, id, id, 1); e != nil || st.changes != 1 {
		t.Fatal("revoked reference trapped", e)
	}
}
