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
	teamhttp "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/http"

	"cloud.google.com/go/firestore"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"

	membershipdb "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/firestore"
	membershiphttp "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/http"
	membershipapp "github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	membership "github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
	orgdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	teamdb "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore"
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore/group"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/team/application/query"
	groupevent "github.com/963sup/line-mini-app-bot/internal/business/team/domain/event"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
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
	groups := &groupquery.Service{Store: store, Membership: member(true)}
	mux := http.NewServeMux()
	invitations := &membershipapp.Service{
		Store: membershipdb.Store{DB: db, Channel: channel},
		Code:  func(_, eventID string) string { return eventID },
	}
	httpapi.RegisterMembership(mux, membershiphttp.Handler{Auth: &owner, Service: invitations})
	httpapi.RegisterOrganizations(mux, Handler{Auth: &owner, Store: orgdb.Store{DB: db, Channel: channel, EnterpriseNavigation: enterprisedb.NavigationFacts}})
	httpapi.RegisterEnterprises(mux, enterprisehttp.Handler{Auth: &owner, Store: enterprisedb.Store{DB: db, Channel: channel, Groups: groups}})
	httpapi.RegisterTeams(mux, teamhttp.Handler{Auth: &owner, Store: teamdb.Store{DB: db, Channel: channel, Groups: groups}})
	call := func(method, path, body string, code int) []byte {
		t.Helper()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, "/api/v1/"+path, strings.NewReader(body))
		r.Header.Set("Idempotency-Key", "workspace-lifecycle")
		mux.ServeHTTP(w, r)
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
	code, err := invitations.Invite(ctx, string(owner), "join-organization", id)
	if err != nil {
		t.Fatal(err)
	}
	if err := invitations.Accept(ctx, other, "accept-organization", code); err != nil {
		t.Fatal(err)
	}
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
	event := groupevent.Observation{GroupID: gid, EventID: "join", At: time.Now(), Name: "Team", Status: "ACTIVE"}
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
	memberTarget := `{"memberId":"` + membership.MemberRef(id, other) + `"}`
	call("PUT", transfer, `{"userId":"`+string(original)+`"}`, 200)
	call("PUT", transfer, `{"memberId":"`+membership.MemberRef(foreignID, other)+`"}`, 400)
	call("PUT", transfer, memberTarget, 204)
	call("PUT", transfer, memberTarget, 404)
	orgPath := "organizations/" + id
	if !strings.Contains(string(call("GET", orgPath+"/members", "", 200)), `"canLeave":true`) {
		t.Fatal("previous owner cannot leave as a member")
	}
	call("DELETE", orgPath+"/membership", "", 204)
	call("GET", orgPath+"/members", "", 404)
	owner = actor(other)
	if !strings.Contains(string(call("GET", orgPath+"/members", "", 200)), `"canRemove":true`) {
		t.Fatal("successor does not have owner capabilities")
	}
}
