package wikihttp_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	repoDB "github.com/963sup/line-mini-app-bot/internal/business/repository/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/wiki"
	wikihttp "github.com/963sup/line-mini-app-bot/internal/business/repository/wiki/http"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
)

type actorAuth struct{ actor string }

func (a *actorAuth) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return a.actor, a.actor, true
}

type members map[string]bool

func (m members) IsMember(_ context.Context, _, user string) (bool, error) { return m[user], nil }

func TestWikiMembershipAndRevisions(t *testing.T) {
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("requires loopback emulator")
	}
	hostname, _, err := net.SplitHostPort(host)
	if err != nil || !net.ParseIP(hostname).IsLoopback() {
		t.Fatal("requires loopback emulator")
	}
	ctx := context.Background()
	client, err := db.NewClient(ctx, "wiki-local")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	channel := client.Collection("channels").NewDoc()
	store := &repoDB.Store{Client: client, ChannelID: channel.ID}
	rid := strings.Repeat("a", 32)
	scope := channel.Collection("repositories").Doc(rid)
	_, err = scope.Set(ctx, map[string]any{"id": rid, "ownerId": "C" + rid, "version": 1, "status": "ACTIVE"})
	if err != nil {
		t.Fatal(err)
	}
	membership := members{"alpha": true, "beta": true}
	auth := &actorAuth{"alpha"}
	mux := http.NewServeMux()
	wikihttp.Register(mux, &wiki.Service{Repositories: &repository.Service{Store: store, Membership: membership}, Store: store}, auth)
	call := func(method, path, body, key string, want int) []byte {
		req := httptest.NewRequest(method, "/api/v1/repositories/"+rid+"/wiki-pages"+path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}
	first := `{"title":"Pump","body":"initial"}`
	key := "create-original-0001"
	var p generated.WikiPage
	if err = json.Unmarshal(call("POST", "", first, key, 200), &p); err != nil {
		t.Fatal(err)
	}
	path := "/" + p.Id
	auth.actor = "beta"
	edit := `{"title":"Pump SOP","body":"revised","expectedVersion":1}`
	call("PATCH", path, edit, "", 200)
	call("PATCH", path, edit, "", 409)
	call("POST", "", `{"title":"pump sop","body":"duplicate"}`, "second-creation-key", 409)
	auth.actor = "alpha"
	if err = json.Unmarshal(call("POST", "", first, key, 200), &p); err != nil || p.Version != 2 || "/"+p.Id != path {
		t.Fatal(p, err)
	}
	var history generated.WikiHistory
	if err = json.Unmarshal(call("GET", path+"/revisions", "", "", 200), &history); err != nil || len(history.Items) != 2 || history.Items[0].AuthorId != "beta" || history.Items[1].Body != "initial" {
		t.Fatal(history, err)
	}
	call("GET", path, "", "", 200)
	call("GET", "", "", "", 200)
	// The same title in a different repository is independent.
	original := rid
	rid = strings.Repeat("b", 32)
	_, err = channel.Collection("repositories").Doc(rid).Set(ctx, map[string]any{"id": rid, "ownerId": "C" + rid, "version": 1, "status": "ACTIVE"})
	if err != nil {
		t.Fatal(err)
	}
	call("POST", "", `{"title":"Pump SOP","body":"second repository"}`, "other-repository-key", 200)
	call("GET", path, "", "", 404)
	rid = original
	delete(membership, "alpha")
	call("GET", path, "", "", 404)
	call("GET", path+"/revisions", "", "", 404)
	call("PATCH", path, edit, "", 404)
}
