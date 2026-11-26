package teamdb

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	membershipdb "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/firestore"
	teamapp "github.com/963sup/line-mini-app-bot/internal/business/team/application"
	teamdomain "github.com/963sup/line-mini-app-bot/internal/business/team/domain"
	groupevent "github.com/963sup/line-mini-app-bot/internal/business/team/domain/event"
	firestorekey "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
)

func TestObservedTeamCanOpenWithLINEAndParentAuthority(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := firestore.NewClient(ctx, "team-activation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	root := db.Collection("channels").NewDoc()
	teamID := "C" + strings.Repeat("a", 32)
	actor := "U" + strings.Repeat("a", 32)
	parent := "a5b7ccea-49f7-4b64-b255-9a79ad22b81f"
	lineMember := false
	store := Store{
		DB: db, Channel: root.ID, OpenMembership: membershipdb.OpenTeamMembership, TeamMember: membershipdb.TeamMember,
	}
	cmd := teamdomain.Activation{Actor: actor, TeamID: teamID, Scope: "enterprise", ManagementID: parent, Maintainer: actor}
	check := func(_ context.Context, group, user string) (bool, error) {
		return lineMember && group == teamID && user == actor, nil
	}
	if err := (LINEObservation{Store: store}).Observe(ctx, groupevent.Observation{GroupID: teamID, Status: "ACTIVE", EventID: "join", At: time.Now()}); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = root.Collection("teams").Doc(teamID).Delete(context.Background()) }()
	enterprise := root.Collection("enterprises").Doc(parent)
	if _, err := enterprise.Set(ctx, map[string]any{"membershipMode": "membership"}); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = enterprise.Delete(context.Background()) }()
	qual := root.Collection("memberships").Doc(firestorekey.Encode("enterprise", parent, actor))
	if _, err := qual.Set(ctx, map[string]any{"status": "active", "role": "owner", "generation": 1, "version": 1}); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = qual.Delete(context.Background()) }()
	if err := teamapp.Activate(ctx, store, check, cmd); !errors.Is(err, teamdomain.ErrActivationDenied) {
		t.Fatalf("actor outside LINE group could open Team: %v", err)
	}
	teamRef := root.Collection("teams").Doc(teamID)
	changeDuringLookup := func(ctx context.Context, _, _ string) (bool, error) {
		_, err := teamRef.Update(ctx, []firestore.Update{{Path: "disabled", Value: true}})
		return true, err
	}
	if err := teamapp.Activate(ctx, store, changeDuringLookup, cmd); !errors.Is(err, teamdomain.ErrActivationDenied) {
		t.Fatalf("disabled during membership check: %v", err)
	}
	if _, err := teamRef.Update(ctx, []firestore.Update{{Path: "disabled", Value: false}}); err != nil {
		t.Fatal(err)
	}
	lineMember = true
	if err := teamapp.Activate(ctx, store, check, cmd); err != nil {
		t.Fatalf("qualified actor could not open observed Team: %v", err)
	}
	defer func() {
		_, _ = root.Collection("memberships").Doc(firestorekey.Encode("team", teamID, actor)).Delete(context.Background())
	}()
	defer func() {
		_, _ = root.Collection("membershipAudits").Doc(firestorekey.Encode("team_opened", teamID)).Delete(context.Background())
	}()
	lineMember = false
	if err := teamapp.Activate(ctx, store, nil, cmd); err != nil {
		t.Fatalf("exact replay required LINE: %v", err)
	}
	if ok, err := store.IsTeamMember(ctx, teamID, actor); err != nil || !ok {
		t.Fatalf("formal qualification missing: %v %v", ok, err)
	}
}
