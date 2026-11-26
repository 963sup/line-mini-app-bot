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
	membershipdb "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/firestore"
	orgdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	repositoryapp "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	labels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
	firestorekey "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
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
	service := repositoryapp.Service{AddressWriter: store}
	if _, err := service.UpdateAddress(ctx, current.OwnerID, current.ID, 1, address, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	persisted, err := store.GetByID(ctx, current.ID)
	if err != nil || persisted.Version != 2 || persisted.Address == nil || *persisted.Address != address {
		t.Fatalf("persisted = %+v, err = %v", persisted, err)
	}
	if _, err := service.UpdateAddress(ctx, current.OwnerID, current.ID, 1, address, now.Add(time.Minute)); !errors.Is(err, repositoryDomainEntity.ErrConflict) {
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
	if err := store.ObserveGroup(ctx, repositoryDomainEntity.GroupObservation{OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: owner, GroupID: group, Name: "工程群組", Status: "ACTIVE", EventID: "join", At: at, Users: []string{owner}}); err != nil {
		t.Fatal(err)
	}
	repository, err := store.GetByGroupID(ctx, group)
	if err != nil || repository.OwnerType != repositoryDomainEntity.OwnerUser || repository.OwnerID != owner || repository.Status != "ACTIVE" || repository.InteractionGroupID != group {
		t.Fatalf("repository = %+v, err = %v", repository, err)
	}
	binding, err := store.lineBindingRef(group).Get(ctx)
	if err != nil || binding.Ref.ID == group {
		t.Fatalf("binding key must not expose external group identity: %v, %v", binding, err)
	}
	var bindingData lineBindingDocument
	if err := binding.DataTo(&bindingData); err != nil || bindingData.GroupID != group || bindingData.RepositoryID != repository.ID {
		t.Fatalf("binding = %+v, err = %v", bindingData, err)
	}
	repositoryData, err := store.repositoryRef(repository.ID).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := repositoryData.Data()["groupId"]; exists {
		t.Fatal("repository duplicates the group binding")
	}
	page, err := store.Candidates(ctx, owner, "", "")
	if err != nil || len(page.Items) != 1 || page.Items[0].InteractionGroupID != group {
		t.Fatalf("candidate binding projection = %+v, %v", page, err)
	}
	if err := store.ObserveGroup(ctx, repositoryDomainEntity.GroupObservation{GroupID: group, Status: "INACTIVE", EventID: "leave", At: at.Add(time.Second)}); err != nil {
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

func TestRepositoryWritesUseCurrentOrganizationQualification(t *testing.T) {
	host, _, err := net.SplitHostPort(os.Getenv("FIRESTORE_EMULATOR_HOST"))
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Skip("requires loopback Firestore emulator")
	}
	ctx := t.Context()
	client, err := db.NewClient(ctx, "repository-qualification-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	root := client.Collection("channels").NewDoc()
	org := root.Collection("organizations").Doc("11111111-1111-4111-8111-111111111111")
	user := "U" + strings.Repeat("b", 32)
	store := &Store{Client: client, ChannelID: root.ID}
	organizations := orgdb.Store{DB: client, Channel: root.ID, OrganizationMember: membershipdb.OrganizationMember}
	store.OrganizationVisible = organizations.VisibleTo
	service := repositoryapp.Service{Store: store, AddressWriter: store, OwnerAccess: organizations}
	const rid = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	repo := root.Collection("repositories").Doc(rid)
	if _, err := repo.Set(ctx, document{ID: rid, OwnerID: org.ID, OwnerType: "ORGANIZATION", Version: 1, Status: "ACTIVE"}); err != nil {
		t.Fatal(err)
	}
	if _, err := org.Set(ctx, map[string]any{"membershipMode": "membership", "ownerUserId": user}); err != nil {
		t.Fatal(err)
	}
	qual := root.Collection("memberships").Doc(firestorekey.Encode("organization", org.ID, user))
	if _, err := qual.Set(ctx, map[string]any{"status": "active", "role": "member", "generation": 1, "version": 1}); err != nil {
		t.Fatal(err)
	}
	address, err := repositoryDomainEntity.NewAddress("TW", "", "臺北")
	if err != nil {
		t.Fatal(err)
	}
	current, err := service.UpdateAddress(ctx, user, rid, 1, address, time.Now())
	if err != nil || current.Version != 2 {
		t.Fatalf("qualified member: %+v %v", current, err)
	}
	if _, err := service.Get(ctx, user, rid); err != nil {
		t.Fatal(err)
	}
	label, err := store.CreateLabel(ctx, current, user, "create-label-receipt", "工程")
	if err != nil {
		t.Fatal(err)
	}
	before, err := repo.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := qual.Delete(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateAddress(ctx, user, rid, current.Version, address, time.Now()); !errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		t.Fatalf("revoked member: %v", err)
	}
	if _, err := store.CreateLabel(ctx, current, user, "create-label-receipt", "工程"); !errors.Is(err, labels.ErrNotFound) {
		t.Fatalf("revoked replay: %v", err)
	}
	if _, err := store.RenameLabel(ctx, current, user, label.ID, "工程", "已改"); !errors.Is(err, labels.ErrNotFound) {
		t.Fatalf("revoked rename: %v", err)
	}
	snap, err := repo.Collection("labels").Doc(label.ID).Get(ctx)
	persisted, err := decodeLabel(snap, err, rid)
	if err != nil || persisted.Name != "工程" {
		t.Fatalf("denied rename changed label: %+v, %v", persisted, err)
	}
	after, err := repo.Get(ctx)
	if err != nil || !after.UpdateTime.Equal(before.UpdateTime) {
		t.Fatalf("denied write changed repository: %v", err)
	}
}
