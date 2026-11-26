package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type bindingStore struct {
	group repositoryDomainEntity.GroupObservation
	err   error
}

func (*bindingStore) ObservePersonal(context.Context, repositoryDomainEntity.PersonalObservation) error {
	return nil
}

func (s *bindingStore) ObserveGroup(_ context.Context, observation repositoryDomainEntity.GroupObservation) error {
	s.group = observation
	return s.err
}

func (*bindingStore) GetByGroupID(context.Context, string) (repositoryDomainEntity.Repository, error) {
	return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
}

func TestBindingsResolveOwnerOnlyForActiveGroup(t *testing.T) {
	store := &bindingStore{}
	resolved := 0
	service := Bindings{Store: store, ResolveOwner: func(context.Context, string) (string, string, error) {
		resolved++
		return string(repositoryDomainEntity.OwnerOrganization), "11111111-1111-4111-8111-111111111111", nil
	}}
	actor := "U" + strings.Repeat("a", 32)
	group := "C" + strings.Repeat("b", 32)
	now := time.Now()
	if err := service.ObserveGroup(t.Context(), actor, group, "工程群組", "ACTIVE", "join", now, []string{actor}); err != nil {
		t.Fatal(err)
	}
	if resolved != 1 || store.group.OwnerType != repositoryDomainEntity.OwnerOrganization || store.group.OwnerID == "" {
		t.Fatalf("resolved=%d observation=%+v", resolved, store.group)
	}
	if err := service.ObserveGroup(t.Context(), "", group, "", "INACTIVE", "leave", now.Add(time.Second), nil); err != nil {
		t.Fatal(err)
	}
	if resolved != 1 || store.group.OwnerID != "" || store.group.OwnerType != "" {
		t.Fatalf("inactive observation resolved owner: %+v", store.group)
	}
}

func TestBindingsPropagateOwnerResolutionFailure(t *testing.T) {
	want := errors.New("owner unavailable")
	store := &bindingStore{}
	service := Bindings{Store: store, ResolveOwner: func(context.Context, string) (string, string, error) {
		return "", "", want
	}}
	err := service.ObserveGroup(t.Context(), "U"+strings.Repeat("a", 32), "C"+strings.Repeat("b", 32), "", "ACTIVE", "message", time.Now(), nil)
	if !errors.Is(err, want) || store.group.EventID != "" {
		t.Fatalf("error=%v observation=%+v", err, store.group)
	}
}
