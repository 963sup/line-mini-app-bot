package enterprisedb_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"
	enterprisehttp "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/http"
	firestorekey "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/google/uuid"
)

type provisioningAuth struct{ actor string }

func (a provisioningAuth) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return a.actor, "", true
}

func TestProvisioningReplayRevalidatesLifecycle(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST")
	}
	ctx := t.Context()
	db, err := firestore.NewClient(ctx, "enterprise-provisioning-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	actor := "U" + strings.Repeat("a", 32)
	for _, tc := range []struct {
		name, mode                            string
		disabled, revoked, otherOwner, scalar bool
		status                                int
	}{
		{name: "existing default legacy", status: 200},
		{name: "explicit legacy", mode: "legacy", status: 200},
		{name: "scalar owner compatibility", scalar: true, status: 200},
		{name: "disabled enterprise", disabled: true, status: 403},
		{name: "formal membership cannot use stale owners", mode: "membership", status: 403},
		{name: "unknown membership mode", mode: "future", status: 403},
		{name: "ownership changed", otherOwner: true, status: 403},
		{name: "entitlement revoked", revoked: true, status: 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := db.Collection("channels").Doc(uuid.NewString())
			id := uuid.NewString()
			owner := actor
			if tc.otherOwner {
				owner = "U" + strings.Repeat("b", 32)
			}
			data := map[string]any{"name": "Existing enterprise", "disabled": tc.disabled, "membershipMode": tc.mode}
			if tc.scalar {
				data["ownerUserId"] = owner
			} else {
				data["ownerUserIds"] = []string{owner}
			}
			enterprise := root.Collection("enterprises").Doc(id)
			grant := root.Collection("enterpriseEntitlements").Doc(firestorekey.Digest([]byte(actor)))
			if _, err := enterprise.Set(ctx, data); err != nil {
				t.Fatal(err)
			}
			if _, err := grant.Set(ctx, map[string]any{"enabled": !tc.revoked, "enterpriseId": id}); err != nil {
				t.Fatal(err)
			}
			before, err := enterprise.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			grantBefore, err := grant.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			handler := enterprisehttp.Handler{Store: enterprisedb.Store{DB: db, Channel: root.ID}, Auth: provisioningAuth{actor}}
			request := httptest.NewRequest(http.MethodPost, "/api/v1/enterprises", strings.NewReader(`{"name":"Retry name"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.CreateEnterprise(response, request)
			if response.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", response.Code, tc.status, response.Body)
			}
			var body struct {
				ID, Name string
				Error    struct{ Code string }
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tc.status == 200 {
				if body.ID != id || body.Name != "Existing enterprise" {
					t.Fatalf("replay changed result: %s", response.Body)
				}
			} else if body.ID != "" || body.Error.Code != "forbidden" {
				t.Fatalf("denial exposed result or changed code: %s", response.Body)
			}
			after, err := enterprise.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			grantAfter, err := grant.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			docs, err := root.Collection("enterprises").Documents(ctx).GetAll()
			if err != nil || len(docs) != 1 || !before.UpdateTime.Equal(after.UpdateTime) || !grantBefore.UpdateTime.Equal(grantAfter.UpdateTime) {
				t.Fatalf("replay mutated resources or entitlement: %v", err)
			}
		})
	}
}
