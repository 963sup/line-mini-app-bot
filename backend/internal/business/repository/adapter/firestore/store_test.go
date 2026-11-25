package firestore

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

func TestUpdateAddressPersistsOnceAndRejectsStaleVersion(t *testing.T) {
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	name, _, err := net.SplitHostPort(host)
	if err != nil || name != "127.0.0.1" && name != "localhost" {
		t.Skip("requires an isolated loopback Firestore emulator")
	}
	ctx := context.Background()
	client, err := db.NewClient(ctx, "repository-address-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	store := &Store{Client: client, ChannelID: "channel"}
	now := time.Now().UTC().Truncate(time.Microsecond)
	current := repositoryDomainEntity.Repository{
		ID: strings.Repeat("a", 32), OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: "U" + strings.Repeat("b", 32),
		Status: "ACTIVE", Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := store.repositoryRef(current.ID).Set(ctx, document{
		ID: current.ID, OwnerType: string(current.OwnerType), OwnerID: current.OwnerID, Status: current.Status, Version: current.Version,
		CreatedAt: current.CreatedAt, UpdatedAt: current.UpdatedAt,
	}); err != nil {
		t.Fatal(err)
	}
	address, err := repositoryDomainEntity.NewAddress("TW", "100", "臺北市中正區")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := current.ChangeAddress(1, address, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpdateAddress(ctx, current, updated); err != nil {
		t.Fatal(err)
	}
	persisted, err := store.GetByID(ctx, current.ID)
	if err != nil || persisted.Version != 2 || persisted.Address == nil || *persisted.Address != address {
		t.Fatalf("persisted = %+v, err = %v", persisted, err)
	}
	if _, err := store.UpdateAddress(ctx, current, updated); !errors.Is(err, repositoryDomainEntity.ErrConflict) {
		t.Fatalf("stale update error = %v", err)
	}
	persisted, err = store.GetByID(ctx, current.ID)
	if err != nil || persisted.Version != 2 || persisted.Address == nil || *persisted.Address != address {
		t.Fatalf("after rejected update = %+v, err = %v", persisted, err)
	}
}

func TestGroupObservationSeparatesOwnerAndBindingLifecycle(t *testing.T) {
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	name, _, err := net.SplitHostPort(host)
	if err != nil || name != "127.0.0.1" && name != "localhost" {
		t.Skip("requires an isolated loopback Firestore emulator")
	}
	ctx := context.Background()
	client, err := db.NewClient(ctx, "repository-binding-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	store := &Store{Client: client, ChannelID: "binding-channel"}
	owner := "U" + strings.Repeat("1", 32)
	group := "C" + strings.Repeat("2", 32)
	at := time.Now().UTC()
	if err := store.Observe(ctx, repositoryDomainEntity.Observation{OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: owner, GroupID: group, Name: "工程群組", Status: "ACTIVE", EventID: "join", At: at, Users: []string{owner}}); err != nil {
		t.Fatal(err)
	}
	repository, err := store.GetByGroupID(ctx, group)
	if err != nil || repository.OwnerType != repositoryDomainEntity.OwnerUser || repository.OwnerID != owner || repository.GroupID != group || repository.Status != "ACTIVE" {
		t.Fatalf("repository = %+v, err = %v", repository, err)
	}
	if err := store.Observe(ctx, repositoryDomainEntity.Observation{OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: owner, GroupID: group, Status: "INACTIVE", EventID: "leave", At: at.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetByGroupID(ctx, group); !errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		t.Fatalf("inactive binding lookup = %v", err)
	}
	persisted, err := store.GetByID(ctx, repository.ID)
	if err != nil || persisted.Status != "ACTIVE" || persisted.OwnerID != owner {
		t.Fatalf("repository lifecycle followed binding: %+v, %v", persisted, err)
	}
}
