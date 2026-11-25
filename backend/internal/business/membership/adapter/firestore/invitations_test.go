package membershipdb

import (
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	"github.com/google/uuid"
)

func TestAcceptRedeliveryPreservesMembershipAndAudit(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST")
	}
	ctx := t.Context()
	db, err := firestore.NewClient(ctx, "demo-membership")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	store := Store{DB: db, Channel: uuid.NewString()}
	root := db.Collection("channels").Doc(store.Channel)
	org := root.Collection("organizations").Doc(uuid.NewString())
	if _, err := org.Create(ctx, map[string]any{"ownerUserId": "owner"}); err != nil {
		t.Fatal(err)
	}
	service := application.Service{Store: store, Code: func(_, event string) string { return event }}
	code, err := service.Invite(ctx, "owner", "issue-event", org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Accept(ctx, "member", "accept-event", code); err != nil {
		t.Fatal(err)
	}
	before, err := org.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Accept(ctx, "member", "accept-event", code); err != nil {
		t.Fatal(err)
	}
	after, err := org.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var record organizationRecord
	if err := after.DataTo(&record); err != nil {
		t.Fatal(err)
	}
	if !before.UpdateTime.Equal(after.UpdateTime) || len(record.Members) != 1 || record.Members[0] != "member" {
		t.Fatalf("unexpected replay write or roster: %v", record.Members)
	}
	audits, err := root.Collection("membershipAudits").Documents(ctx).GetAll()
	if err != nil || len(audits) != 1 {
		t.Fatalf("audit count=%d err=%v", len(audits), err)
	}
}
