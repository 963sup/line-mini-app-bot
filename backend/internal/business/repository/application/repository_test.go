package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type candidatePage struct {
	port.Store
	dto.Page
}

func (p candidatePage) Candidates(context.Context, string, string, string) (dto.Page, error) {
	return p.Page, nil
}

func TestListForOwnerPreservesAuthorizationAndContinuation(t *testing.T) {
	user := "U" + strings.Repeat("a", 32)
	service := Service{Store: candidatePage{Page: dto.Page{Items: []entity.Repository{
		{ID: "group", OwnerType: entity.OwnerUser, OwnerID: user, Status: "ACTIVE"}, {ID: "inactive", OwnerType: entity.OwnerUser, OwnerID: user, Status: "INACTIVE"}, {ID: "personal", OwnerType: entity.OwnerUser, OwnerID: user, Status: "ACTIVE"},
	}, NextCursor: "more"}}}
	page, err := service.List(t.Context(), user, "", user)
	if err != nil || len(page.Items) != 2 || page.Items[0].ID != "group" || page.Items[1].ID != "personal" || page.NextCursor != "more" {
		t.Fatalf("unexpected page: %+v, %v", page, err)
	}
	page, err = service.List(t.Context(), "U"+strings.Repeat("c", 32), "", user)
	if err != nil || len(page.Items) != 0 || page.NextCursor != "more" {
		t.Fatalf("unexpected page for another actor: %+v, %v", page, err)
	}
}

type singleRepository struct {
	port.Store
	item    entity.Repository
	updated entity.Repository
}

func (s singleRepository) GetByID(context.Context, string) (entity.Repository, error) {
	return s.item, nil
}

func (s *singleRepository) RunAddressChange(_ context.Context, _ string, change port.AddressChange) (entity.Repository, error) {
	updated, err := change(s.item, nil)
	if err == nil {
		s.updated = updated
	}
	return updated, err
}

func TestGetRejectsInactiveRepositoryForItsOwner(t *testing.T) {
	item := entity.Repository{ID: strings.Repeat("a", 32), OwnerType: entity.OwnerUser, OwnerID: "U" + strings.Repeat("b", 32), Status: "INACTIVE"}
	service := Service{Store: singleRepository{item: item}}
	if _, err := service.Get(t.Context(), item.OwnerID, item.ID); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("inactive repository: got %v, want not found", err)
	}
}

func TestUpdateAddressUsesAuthorizedRepositoryAndVersion(t *testing.T) {
	owner := "U" + strings.Repeat("b", 32)
	item := entity.Repository{ID: strings.Repeat("a", 32), OwnerType: entity.OwnerUser, OwnerID: owner, Status: "ACTIVE", Version: 3}
	store := &singleRepository{item: item}
	service := Service{Store: store, AddressWriter: store}
	address, err := entity.NewAddress(" tw ", " 100 ", " 臺北市中正區 ")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.UpdateAddress(t.Context(), owner, item.ID, item.Version, address, time.Now().UTC())
	if err != nil || updated.Version != 4 || updated.Address == nil || updated.Address.CountryCode != "TW" || store.updated.Address == nil {
		t.Fatalf("updated = %+v, persisted = %+v, err = %v", updated, store.updated, err)
	}
	if _, err := service.UpdateAddress(t.Context(), owner, item.ID, item.Version-1, address, time.Now().UTC()); !errors.Is(err, entity.ErrConflict) {
		t.Fatalf("stale update = %v", err)
	}
}
