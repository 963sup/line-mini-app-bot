package organizationhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	enterprisehttp "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/http"
	teamhttp "github.com/963sup/line-mini-app-bot/internal/business/team/http"

	"cloud.google.com/go/firestore"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/group"
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/firestore"
	orgdb "github.com/963sup/line-mini-app-bot/internal/business/organization/firestore"
	teamdb "github.com/963sup/line-mini-app-bot/internal/business/team/firestore"
	"github.com/963sup/line-mini-app-bot/internal/httpapi"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"github.com/google/uuid"
)

type actor string

func (a *actor) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return string(*a), "", true
}

type member bool

func (m member) IsMember(context.Context, string, string) (bool, error) { return bool(m), nil }
func TestWorkspaceLifecycle(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires local Firestore emulator")
	}
	ctx := t.Context()
	db, err := firestore.NewClient(ctx, "org-test-local")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if e := db.Close(); e != nil {
			t.Error(e)
		}
	})
	channel := uuid.NewString()
	owner := actor("U" + strings.Repeat("a", 32))
	other := "U" + strings.Repeat("b", 32)
	store := &groupdb.Store{Client: db, ChannelID: channel}
	groups := &group.Service{Store: store, Membership: member(true)}
	mux := http.NewServeMux()
	httpapi.RegisterOrganizations(mux, Handler{Auth: &owner, Store: orgdb.Store{DB: db, Channel: channel}})
	httpapi.RegisterEnterprises(mux, enterprisehttp.Handler{Auth: &owner, Store: enterprisedb.Store{DB: db, Channel: channel, Groups: groups}})
	httpapi.RegisterTeams(mux, teamhttp.Handler{Auth: &owner, Store: teamdb.Store{DB: db, Channel: channel, Groups: groups}})
	call := func(method, path, body string, code int) []byte {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, "/api/v1/"+path, strings.NewReader(body)))
		if w.Code != code {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}
	var org generated.Organization
	if err = json.Unmarshal(call("POST", "organizations", `{"name":"Taipei"}`, 200), &org); err != nil {
		t.Fatal(err)
	}
	root := db.Collection("channels").Doc(channel)
	id := org.OrganizationId
	var provisioned generated.EnterpriseLink
	gid := "C" + strings.Repeat("a", 32)
	save := func(ref *firestore.DocumentRef, data map[string]any) {
		t.Helper()
		if _, e := ref.Set(ctx, data, firestore.MergeAll); e != nil {
			t.Fatal(e)
		}
	}
	call("POST", "enterprises", `{"name":"HQ"}`, 403)
	save(root.Collection("enterpriseEntitlements").Doc(firestorekey.Digest([]byte(owner))), map[string]any{"enabled": true})
	if !strings.Contains(string(call("GET", "me/organization-context", "", 200)), `"canCreateEnterprise":true`) {
		t.Fatal("eligible user cannot create")
	}
	if err = json.Unmarshal(call("POST", "enterprises", `{"name":"HQ"}`, 200), &provisioned); err != nil {
		t.Fatal(err)
	}
	enterprise := provisioned.Id
	ep := "enterprises/" + enterprise
	if !strings.Contains(string(call("POST", "enterprises", `{"name":"HQ"}`, 200)), enterprise) {
		t.Fatal("duplicate enterprise on retry")
	}
	if !strings.Contains(string(call("GET", "me/organization-context", "", 200)), enterprise) {
		t.Fatal("empty enterprise not discoverable")
	}
	original := owner
	owner = actor(other)
	call("GET", ep, "", 404)
	call("POST", ep+"/owners", `{"userId":"`+other+`"}`, 404)
	owner = original
	call("POST", ep+"/organizations", `{"organizationId":"`+id+`"}`, 200)
	save(root.Collection("organizations").Doc(id), map[string]any{"memberUserIds": []string{other}})
	owner = actor(other)
	if !strings.Contains(string(call("GET", ep, "", 200)), `"canManage":false`) {
		t.Fatal("member acquired owner privileges")
	}
	owner = original
	call("POST", ep+"/owners", `{"userId":"`+other+`"}`, 200)
	owner = actor(other)
	if !strings.Contains(string(call("GET", ep, "", 200)), `"canManage":true`) {
		t.Fatal("additional owner cannot manage")
	}
	owner = original
	selection := `{"organizationId":"` + id + `"}`
	call("PUT", "me/organization-context", selection, 200)
	call("PUT", "me/organization-context", `{"organizationId":"`+uuid.NewString()+`"}`, 404)
	if !strings.Contains(string(call("GET", "me/organization-context", "", 200)), `"groupLabel":"HQ"`) {
		t.Fatal("missing selected enterprise")
	}
	event := group.Observation{GroupID: gid, EventID: "join", At: time.Now(), Name: "Team", Status: "ACTIVE"}
	if err = store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	mgmt := "teams/" + gid + "/management"
	payload := `{"managementScope":"enterprise","managementId":"` + enterprise + `"}`
	call("PUT", mgmt, payload, 404)
	save(root.Collection("teams").Doc(gid), map[string]any{"managerUserIds": []string{string(owner)}})
	call("PUT", mgmt, payload, 204)
	if err = store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	link := "organizations/" + id + "/enterprise-teams/" + gid
	foreignID := uuid.NewString()
	save(root.Collection("organizations").Doc(foreignID), map[string]any{"enterpriseId": uuid.NewString(), "name": "Other", "ownerUserId": string(owner)})
	call("PUT", "organizations/"+foreignID+"/enterprise-teams/"+gid, "", 404)
	call("PUT", mgmt, `{"managementScope":"organization","managementId":"`+id+`"}`, 409)
	call("PUT", link, "", 204)
	call("PUT", link, "", 204)
	viewer := actor("U" + strings.Repeat("c", 32))
	save(root.Collection("organizations").Doc(id), map[string]any{"memberUserIds": []string{other, string(viewer)}})
	owner = viewer
	if !strings.Contains(string(call("GET", ep, "", 200)), gid) {
		t.Fatal("assigned team missing for member")
	}
	groups.Membership = member(false)
	if strings.Contains(string(call("GET", ep, "", 200)), gid) {
		t.Fatal("enterprise exposed team after LINE membership revoked")
	}
	groups.Membership = member(true)
	owner = original
	list := "organizations/" + id + "/teams"
	if !strings.Contains(string(call("GET", list, "", 200)), gid) {
		t.Fatal("team missing")
	}
	groups.Membership = member(false)
	if strings.Contains(string(call("GET", list, "", 200)), gid) {
		t.Fatal("membership revoked")
	}
	groups.Membership = member(true)
	call("DELETE", link, "", 204)
	if strings.Contains(string(call("GET", list, "", 200)), gid) {
		t.Fatal("assignment not revoked")
	}
	transfer := "organizations/" + id + "/owner"
	call("PUT", transfer, `{"userId":"`+other+`"}`, 200)
	call("PUT", transfer, `{"userId":"`+string(owner)+`"}`, 404)
}
