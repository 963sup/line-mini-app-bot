package wikihttp_test

import (
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/application/wiki"
)

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"

	repoDB "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore"

	wikihttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/wiki"
	firestorekey "github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
)

type actorAuth struct{ actor string }

func (a *actorAuth) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return a.actor, a.actor, true
}

type members map[string]bool

func (m members) CanAccessOwner(_ context.Context, _, _, user string) (bool, error) {
	return m[user], nil
}

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
	_, err = scope.Set(ctx, map[string]any{"id": rid, "ownerType": "ORGANIZATION", "ownerId": "11111111-1111-4111-8111-111111111111", "groupId": "C" + rid, "version": 1, "status": "ACTIVE"})
	if err != nil {
		t.Fatal(err)
	}
	membership := members{"alpha": true, "beta": true}
	auth := &actorAuth{"alpha"}
	mux := http.NewServeMux()
	wikihttp.Register(mux, &repositoryApplicationWiki.Service{Repositories: &repositoryApplication.Service{Store: store, OwnerAccess: membership}, Store: store}, auth)
	request := func(method, path, body, key string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1/repositories/"+rid+"/wiki-pages"+path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}
	call := func(method, path, body, key string, want int) []byte {
		t.Helper()
		w := request(method, path, body, key)
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
	for _, tc := range []struct {
		body []byte
		want generated.APIErrorErrorConflictReason
	}{
		{call("PATCH", path, edit, "", 409), generated.StaleVersion},
		{call("POST", "", `{"title":"pump sop","body":"duplicate"}`, "second-creation-key", 409), generated.DuplicateTitle},
	} {
		var result generated.APIError
		if err := json.Unmarshal(tc.body, &result); err != nil || result.Error.Code != "conflict" || result.Error.ConflictReason == nil || *result.Error.ConflictReason != tc.want {
			t.Fatalf("want conflict %s, got %s", tc.want, tc.body)
		}
	}
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
	t.Run("concurrent title reservation", func(t *testing.T) {
		start := make(chan struct{})
		results := make(chan int, 2)
		for _, key := range []string{"race-creation-key-one", "race-creation-key-two"} {
			go func() {
				<-start
				results <- request("POST", "", `{"title":"Concurrent","body":"test"}`, key).Code
			}()
		}
		close(start)
		first, second := <-results, <-results
		if min(first, second) != http.StatusOK || max(first, second) != http.StatusConflict {
			t.Fatalf("concurrent responses = %d, %d", first, second)
		}
	})

	t.Run("legacy document and failed revision commit", func(t *testing.T) {
		id := strings.Repeat("c", 32)
		doc := scope.Collection("wikiPages").Doc(id)
		legacy := map[string]any{
			"id": id, "title": "Legacy", "body": "stored body", "authorId": "beta",
			"version": int64(1), "updatedAt": time.Unix(1700000000, 0).UTC(),
		}
		if _, err := doc.Set(ctx, legacy); err != nil {
			t.Fatal(err)
		}
		titleIndex := scope.Collection("wikiTitles").Doc(firestorekey.Encode("legacy"))
		if _, err := titleIndex.Set(ctx, map[string]any{"pageId": id}); err != nil {
			t.Fatal(err)
		}
		legacyPath := "/" + id
		var page generated.WikiPage
		if err := json.Unmarshal(call("GET", legacyPath, "", "", 200), &page); err != nil ||
			page.Id != id || page.Title != "Legacy" || page.Body != "stored body" || page.AuthorId != "beta" || page.Version != 1 {
			t.Fatalf("legacy page = %+v, %v", page, err)
		}
		// Force the final revision Create to fail after the page and title writes are queued.
		revision := doc.Collection("revisions").Doc("0000000000000002")
		if _, err := revision.Set(ctx, map[string]any{"sentinel": true}); err != nil {
			t.Fatal(err)
		}
		response := request("PATCH", legacyPath, `{"title":"Rollback title","body":"new","expectedVersion":1}`, "")
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected failed commit, got %d: %s", response.Code, response.Body)
		}
		if err := json.Unmarshal(call("GET", legacyPath, "", "", 200), &page); err != nil ||
			page.Title != "Legacy" || page.Body != "stored body" || page.Version != 1 {
			t.Fatalf("page changed after failed commit: %+v, %v", page, err)
		}
		snapshot, err := revision.Get(ctx)
		if err != nil || snapshot.Data()["sentinel"] != true {
			t.Fatalf("existing revision changed: %v", err)
		}
		call("POST", "", `{"title":"Legacy","body":"reserved"}`, "legacy-duplicate-key", 409)
		call("POST", "", `{"title":"Rollback title","body":"available"}`, "rollback-creation-key", 200)
	})

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
