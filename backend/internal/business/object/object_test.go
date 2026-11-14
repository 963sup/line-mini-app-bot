package object

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

func content() Content {
	return Content{Name: "工項", Description: "完整說明", Type: "設備", Quantity: "11.000000", Unit: "SET", SourceDocument: "doc", SourceRevision: "v0", SourceItemNo: "10"}
}

func TestQuantityAndContent(t *testing.T) {
	c, e := Normalize(content())
	if e != nil || c.Quantity != "11" {
		t.Fatalf("canonical: %+v %v", c, e)
	}
	for _, v := range []string{"0", "0.000000", "-1", "01", "1e3", "1.0000001", "1000000000000000000", "NaN"} {
		c := content()
		c.Quantity = v
		if _, e := Normalize(c); !errors.Is(e, ErrInput) {
			t.Errorf("accepted %q", v)
		}
	}
	for _, v := range []string{"0.000001", "999999999999999999.999999"} {
		c := content()
		c.Quantity = v
		if _, e := Normalize(c); e != nil {
			t.Errorf("rejected %q: %v", v, e)
		}
	}
	c = content()
	c.Name = strings.Repeat("項", 121)
	if _, e := Normalize(c); !errors.Is(e, ErrInput) {
		t.Fatal("oversize accepted")
	}
}

func TestUpdateAuthorityVersionAndProvenance(t *testing.T) {
	c, _ := Normalize(content())
	current := Object{Content: c, CreatorID: "A", Status: "not_started", Version: 1}
	u := Update{Content: c, Status: "completed", ExpectedVersion: 1}
	u.SourceDocument = ""
	u.SourceRevision = ""
	u.SourceItemNo = ""
	if _, e := ApplyUpdate(current, "B", u); !errors.Is(e, ErrForbidden) {
		t.Fatal(e)
	}
	u.ExpectedVersion = 2
	if _, e := ApplyUpdate(current, "A", u); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	u.ExpectedVersion = 1
	next, e := ApplyUpdate(current, "A", u)
	if e != nil || next.Status != "completed" || next.Version != 2 || next.SourceDocument != "doc" || next.SourceRevision != "v0" || next.SourceItemNo != "10" {
		t.Fatalf("source omission/status: %+v %v", next, e)
	}
	empty := ""
	u.Sources.SourceDocument = &empty
	next, e = ApplyUpdate(current, "A", u)
	if e != nil || next.SourceDocument != "" || next.SourceRevision != "v0" {
		t.Fatalf("explicit clear: %+v %v", next, e)
	}
	u = Update{Content: c, Status: current.Status, ExpectedVersion: 1}
	next, e = ApplyUpdate(current, "A", u)
	if e != nil || next.Version != 1 {
		t.Fatal("noop", e)
	}
	u.Status = "approved"
	if _, e = ApplyUpdate(current, "A", u); !errors.Is(e, ErrInput) {
		t.Fatal(e)
	}
}

type repositories struct {
	r repository.Repository
	e error
}

func (r repositories) Get(context.Context, string, string) (repository.Repository, error) {
	return r.r, r.e
}

type stubStore struct {
	calls int
	scope string
}

func (s *stubStore) List(_ context.Context, r repository.Repository, _, _ string) (Page, error) {
	s.calls++
	s.scope = r.ID
	return Page{}, nil
}

func (s *stubStore) Get(_ context.Context, r repository.Repository, _ string) (Object, error) {
	s.calls++
	s.scope = r.ID
	return Object{}, nil
}

func (s *stubStore) Create(_ context.Context, r repository.Repository, _, _ string, _ Content) (Object, error) {
	s.calls++
	s.scope = r.ID
	return Object{}, nil
}

func (s *stubStore) Update(_ context.Context, r repository.Repository, _, _ string, _ Update) (Object, error) {
	s.calls++
	s.scope = r.ID
	return Object{}, nil
}

func TestServiceScopesEveryOperation(t *testing.T) {
	id := strings.Repeat("a", 32)
	r := repository.Repository{ID: id, GroupID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}
	store := &stubStore{}
	s := Service{Repositories: repositories{r: r}, Store: store}
	ctx := context.Background()
	if _, e := s.Create(ctx, "A", id, strings.Repeat("k", 16), content()); e != nil {
		t.Fatal(e)
	}
	if store.scope != id || store.calls != 1 {
		t.Fatal(store)
	}
	for _, rr := range []repositories{{r: r, e: repository.ErrNotFound}, {r: repository.Repository{ID: id, Status: "INACTIVE"}}, {r: repository.Repository{ID: strings.Repeat("b", 32), GroupID: r.GroupID, Status: "ACTIVE"}}} {
		s.Repositories = rr
		before := store.calls
		_, e1 := s.List(ctx, "A", id, "")
		_, e2 := s.Get(ctx, "A", id, id)
		_, e3 := s.Create(ctx, "A", id, strings.Repeat("k", 16), content())
		_, e4 := s.Update(ctx, "A", id, id, Update{})
		for _, e := range []error{e1, e2, e3, e4} {
			if !errors.Is(e, ErrNotFound) {
				t.Error(e)
			}
		}
		if store.calls != before {
			t.Fatal("unauthorized store call")
		}
	}
}
