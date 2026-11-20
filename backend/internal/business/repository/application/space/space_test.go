package space

import (
	"context"
	"errors"
	"strings"
	"testing"

	repositoryApplicationDtoSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/space"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntitySpace "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/space"
)

type repoFake struct {
	r   repositoryDomainEntity.Repository
	err error
}

func (f repoFake) Get(context.Context, string, string) (repositoryDomainEntity.Repository, error) {
	return f.r, f.err
}

type storeFake struct {
	calls int
	name  string
}

func (f *storeFake) List(context.Context, repositoryDomainEntity.Repository, string, string) (repositoryApplicationDtoSpace.Page, error) {
	f.calls++
	return repositoryApplicationDtoSpace.Page{}, nil
}

func (f *storeFake) Get(context.Context, repositoryDomainEntity.Repository, string) (repositoryDomainEntitySpace.Space, error) {
	f.calls++
	return repositoryDomainEntitySpace.Space{}, nil
}

func (f *storeFake) Create(_ context.Context, _ repositoryDomainEntity.Repository, _, _, n string) (repositoryDomainEntitySpace.Space, error) {
	f.calls++
	f.name = n
	return repositoryDomainEntitySpace.Space{Name: n}, nil
}

func TestServiceAuthorityAndInput(t *testing.T) {
	r := repositoryDomainEntity.Repository{ID: strings.Repeat("a", 32), OwnerID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}
	for _, tc := range []struct {
		name, actor, key, title string
		repo                    repositoryDomainEntity.Repository
		repoErr, errorWant      error
	}{
		{"valid", "actor", "creation-key-12345", "  一樓  ", r, nil, nil},
		{"missing actor", "", "creation-key-12345", "name", r, nil, repositoryDomainEntitySpace.ErrNotFound},
		{"not member", "actor", "creation-key-12345", "name", r, repositoryDomainEntity.ErrNotFound, repositoryDomainEntitySpace.ErrNotFound},
		{"inactive", "actor", "creation-key-12345", "name", repositoryDomainEntity.Repository{ID: r.ID, OwnerID: r.OwnerID, Status: "INACTIVE"}, nil, repositoryDomainEntitySpace.ErrNotFound},
		{"blank", "actor", "creation-key-12345", " ", r, nil, repositoryDomainEntitySpace.ErrInput},
		{"long", "actor", "creation-key-12345", strings.Repeat("字", 121), r, nil, repositoryDomainEntitySpace.ErrInput},
		{"key", "actor", "short", "name", r, nil, repositoryDomainEntitySpace.ErrInput},
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
