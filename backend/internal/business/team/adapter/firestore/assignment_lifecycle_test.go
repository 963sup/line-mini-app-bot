package teamdb

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	organization "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
	teamapp "github.com/963sup/line-mini-app-bot/internal/business/team/application"
)

func TestTeamWritesRequireActiveLifecycle(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	db, err := firestore.NewClient(ctx, "team-lifecycle-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	root := db.Collection("channels").NewDoc()
	orgID, entID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	actor, teamID := "owner", "C"+strings.Repeat("a", 32)
	for collection, id := range map[string]string{"organizations": orgID, "enterprises": entID} {
		if _, err = root.Collection(collection).Doc(id).Set(ctx, map[string]any{"ownerUserId": actor, "enterpriseId": entID}); err != nil {
			t.Fatal(err)
		}
	}
	store := Store{DB: db, Channel: root.ID}
	ref := root.Collection("teams").Doc(teamID)
	link := root.Collection("organizations").Doc(orgID).Collection("enterpriseTeams").Doc(teamID)
	for _, tc := range []struct {
		status   string
		disabled bool
		allowed  bool
	}{{"ACTIVE", false, true}, {"INACTIVE", false, false}, {"ACTIVE", true, false}, {"", false, false}} {
		t.Run(tc.status+map[bool]string{true: "-disabled"}[tc.disabled], func(t *testing.T) {
			_, err := ref.Set(ctx, map[string]any{"status": tc.status, "disabled": tc.disabled, "managerUserIds": []string{actor}, "managementScope": "enterprise", "managementId": entID})
			if err != nil {
				t.Fatal(err)
			}
			err = store.Manage(ctx, actor, teamID, "enterprise", entID)
			if tc.allowed && err != nil || !tc.allowed && !errors.Is(err, organization.ErrNotFound) {
				t.Fatalf("management lifecycle: %v", err)
			}
			for _, assigned := range []bool{true, false} {
				err = teamapp.AssignEnterpriseTeam(ctx, store, actor, orgID, teamID, assigned)
				if tc.allowed && err != nil || !tc.allowed && !errors.Is(err, teamapp.ErrAssignmentNotFound) {
					t.Fatalf("assignment %v lifecycle: %v", assigned, err)
				}
			}
			if _, err := link.Get(ctx); err == nil {
				t.Fatal("removed or rejected assignment unexpectedly persists")
			}
		})
	}
}
