package teamdb

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
	groupevent "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGroupCatalogUsesFormalTeamQualification(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := firestore.NewClient(ctx, "team-catalog-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	root := client.Collection("channels").NewDoc()
	id := "C" + strings.Repeat("a", 32)
	team := root.Collection("teams").Doc(id)
	candidate := root.Collection("memberships").NewDoc()
	defer func() { _, _ = team.Delete(context.Background()) }()
	defer func() { _, _ = candidate.Delete(context.Background()) }()
	if _, err := candidate.Set(ctx, map[string]any{"userId": "actor", "resourceType": "team", "resourceId": id}); err != nil {
		t.Fatal(err)
	}
	if _, err := team.Set(ctx, map[string]any{"name": "Team", "status": "ACTIVE", "membershipMode": "membership"}); err != nil {
		t.Fatal(err)
	}
	allowed := true
	catalog := GroupCatalog{Store: Store{DB: client, Channel: root.ID, TeamMember: func(*firestore.Transaction, *firestore.DocumentRef, string, string, string, string) (bool, error) {
		return allowed, nil
	}}}
	page, err := catalog.List(ctx, "actor", "")
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != id || page.Items[0].Name != "Team" {
		t.Fatalf("formal team missing: %+v %v", page, err)
	}
	allowed = false
	page, err = catalog.List(ctx, "actor", "")
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("candidate bypassed formal membership: %+v %v", page, err)
	}
	if _, err := catalog.Get(ctx, "actor", id); !errors.Is(err, groupentity.ErrNotFound) {
		t.Fatalf("detail did not deny revoked membership: %v", err)
	}
}

func TestLINEObservationWritesOnlyTeamAndRejectsOlderEvents(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := firestore.NewClient(ctx, "team-observation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	root := client.Collection("channels").NewDoc()
	id := "C" + strings.Repeat("a", 32)
	store := LINEObservation{Store: Store{DB: client, Channel: root.ID}}
	ref := root.Collection("teams").Doc(id)
	defer func() { _, _ = ref.Delete(context.Background()) }()
	legacy, err := ref.Set(ctx, map[string]any{"status": "INACTIVE", "name": "Existing Team"})
	if err != nil {
		t.Fatal(err)
	}
	stale := groupevent.Observation{GroupID: id, Status: "ACTIVE", EventID: "old-redelivery", At: legacy.UpdateTime.Add(-time.Second)}
	if err := store.Observe(ctx, stale); err != nil {
		t.Fatal(err)
	}
	before, err := ref.Get(ctx)
	if err != nil || before.Data()["status"] != "INACTIVE" || before.Data()["lineEventId"] != nil {
		t.Fatalf("first delayed event changed a legacy Team: %v %v", before, err)
	}
	at := legacy.UpdateTime.Add(time.Second)
	event := groupevent.Observation{GroupID: id, Status: "ACTIVE", EventID: "join", Name: "Team", At: at, Users: []string{"actor"}}
	if err := store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	if accepted, err := ref.Get(ctx); err != nil || accepted.Data()["lineEventId"] != "join" {
		t.Fatalf("fresh event did not establish its watermark: %v %v", accepted, err)
	}
	if _, err := ref.Update(ctx, []firestore.Update{{Path: "managementId", Value: "owner"}}); err != nil {
		t.Fatal(err)
	}
	leave := event
	leave.EventID, leave.Status, leave.At = "leave", "INACTIVE", at.Add(time.Second)
	if err := store.Observe(ctx, leave); err != nil {
		t.Fatal(err)
	}
	if err := store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	team, err := ref.Get(ctx)
	if err != nil || team.Data()["status"] != "INACTIVE" || team.Data()["managementId"] != "owner" {
		t.Fatalf("observation changed accepted state or ownership: %v %v", team, err)
	}
	if _, err := root.Collection("groups").Doc(id).Get(ctx); status.Code(err) != codes.NotFound {
		t.Fatalf("old group document was written: %v", err)
	}
}
