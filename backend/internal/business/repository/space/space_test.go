package space

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/repository"
)

type repoFake struct {
	r   repository.Repository
	err error
}

func (f repoFake) Get(context.Context, string, string) (repository.Repository, error) {
	return f.r, f.err
}

type storeFake struct {
	calls int
	name  string
}

func (f *storeFake) List(context.Context, repository.Repository, string, string) (Page, error) {
	f.calls++
	return Page{}, nil
}

func (f *storeFake) Get(context.Context, repository.Repository, string) (Space, error) {
	f.calls++
	return Space{}, nil
}

func (f *storeFake) Create(_ context.Context, _ repository.Repository, _, _, n string) (Space, error) {
	f.calls++
	f.name = n
	return Space{Name: n}, nil
}

func TestServiceAuthorityAndInput(t *testing.T) {
	r := repository.Repository{ID: strings.Repeat("a", 32), GroupID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}
	for _, tc := range []struct {
		name, actor, key, title string
		repo                    repository.Repository
		repoErr, errorWant      error
	}{
		{"valid", "actor", "creation-key-12345", "  一樓  ", r, nil, nil},
		{"missing actor", "", "creation-key-12345", "name", r, nil, ErrNotFound},
		{"not member", "actor", "creation-key-12345", "name", r, repository.ErrNotFound, ErrNotFound},
		{"inactive", "actor", "creation-key-12345", "name", repository.Repository{ID: r.ID, GroupID: r.GroupID, Status: "INACTIVE"}, nil, ErrNotFound},
		{"blank", "actor", "creation-key-12345", " ", r, nil, ErrInput},
		{"long", "actor", "creation-key-12345", strings.Repeat("字", 121), r, nil, ErrInput},
		{"key", "actor", "short", "name", r, nil, ErrInput},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &storeFake{}
			s := Service{Repositories: repoFake{tc.repo, tc.repoErr}, Store: f}
			_, e := s.Create(context.Background(), tc.actor, r.ID, tc.key, tc.title)
			if !errors.Is(e, tc.errorWant) {
				t.Fatalf("error=%v want %v", e, tc.errorWant)
			}
			if tc.errorWant != nil && f.calls != 0 {
				t.Fatal("rejected request reached storage")
			}
			if tc.errorWant == nil && f.name != "一樓" {
				t.Fatal("name not normalized")
			}
		})
	}
}
