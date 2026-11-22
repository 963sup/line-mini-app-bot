package organizationdb

import (
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"
	dto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/google/uuid"
)

func TestCreationConversation(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires isolated emulator")
	}
	ctx := t.Context()
	db, e := firestore.NewClient(ctx, "org-chat-test")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := db.Close(); e != nil {
			t.Error(e)
		}
	})
	s := Store{DB: db, Channel: uuid.NewString(), CanCreateEnterprise: enterprisedb.CanStartCreation, CreateEnterprise: enterprisedb.CreateFromConversation}
	root := db.Collection("channels").Doc(s.Channel)
	user := "U" + strings.Repeat("a", 32)
	at := time.Now().Add(-time.Minute)
	send := func(key, text string) *dto.CreationReply {
		t.Helper()
		at = at.Add(time.Millisecond)
		r, e := s.Conversation(ctx, user, key, text, at)
		if e != nil {
			t.Fatal(e)
		}
		return r
	}
	send("start", "建立組織")
	if r := send("cancel", "取消"); !strings.Contains(r.Text, "取消") {
		t.Fatal(r)
	}
	if send("unrelated", "名稱") != nil {
		t.Fatal("cancelled prompt consumed message")
	}
	send("second", "建立組織")
	r := send("name", "台北")
	if r.ID == "" || r.Kind != "organization" {
		t.Fatal(r)
	}
	if replay := send("name", "台北"); *replay != *r {
		t.Fatal(replay)
	}
	if docs, e := root.Collection("organizations").Documents(ctx).GetAll(); e != nil || len(docs) != 1 {
		t.Fatal("duplicate resource", e)
	}
	send("timeout", "建立組織")
	ref := root.Collection("Conversations").Doc(firestorekey.Digest([]byte(user)))
	if _, e := ref.Update(ctx, []firestore.Update{{Path: "Expires", Value: time.Now().Add(-time.Second)}}); e != nil {
		t.Fatal(e)
	}
	if r := send("late", "逾時名稱"); !strings.Contains(r.Text, "逾時") {
		t.Fatal(r)
	}
	grant := root.Collection("enterpriseEntitlements").Doc(firestorekey.Digest([]byte(user)))
	allow := func(enabled bool) {
		if _, e := grant.Set(ctx, map[string]any{"enabled": enabled}); e != nil {
			t.Fatal(e)
		}
	}
	if r := send("denied", "建立企業"); !strings.Contains(r.Text, "資格") {
		t.Fatal(r)
	}
	allow(true)
	send("allowed", "建立企業")
	allow(false)
	if r := send("revoked", "HQ"); r.ID != "" || !strings.Contains(r.Text, "資格") {
		t.Fatal(r)
	}
	allow(true)
	send("restart", "建立企業")
	if r := send("hq", "HQ"); r.Kind != "enterprise" || r.ID == "" {
		t.Fatal(r)
	}
}
