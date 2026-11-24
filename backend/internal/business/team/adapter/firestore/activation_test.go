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
	organization "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
	groupevent "github.com/963sup/line-mini-app-bot/internal/business/team/domain/event"
	firestorekey "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
)

type activationMembership func(context.Context, string, string) (bool, error)

func (f activationMembership) IsMember(ctx context.Context, group, user string) (bool, error) {
	return f(ctx, group, user)
}

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
		LINEActivationMember: activationMembership(func(_ context.Context, group, user string) (bool, error) {
			return lineMember && group == teamID && user == actor, nil
		}),
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
	if err := store.Activate(ctx, actor, teamID, "enterprise", parent, actor); !errors.Is(err, organization.ErrNotFound) {
		t.Fatalf("actor outside LINE group could open Team: %v", err)
	}
	lineMember = true
	if err := store.Activate(ctx, actor, teamID, "enterprise", parent, actor); err != nil {
		t.Fatalf("qualified actor could not open observed Team: %v", err)
	}
	defer func() {
		_, _ = root.Collection("memberships").Doc(firestorekey.Encode("team", teamID, actor)).Delete(context.Background())
	}()
	defer func() {
		_, _ = root.Collection("membershipAudits").Doc(firestorekey.Encode("team_opened", teamID)).Delete(context.Background())
	}()
	lineMember = false
	if ok, err := store.IsTeamMember(ctx, teamID, actor); err != nil || !ok {
		t.Fatalf("formal qualification missing: %v %v", ok, err)
	}
}
