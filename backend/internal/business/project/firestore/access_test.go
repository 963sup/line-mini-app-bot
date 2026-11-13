package firestore_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	p "github.com/963sup/line-mini-app-bot/internal/business/project"
	projectDB "github.com/963sup/line-mini-app-bot/internal/business/project/firestore"
	projectHTTP "github.com/963sup/line-mini-app-bot/internal/business/project/http"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	repoDB "github.com/963sup/line-mini-app-bot/internal/business/repository/firestore"
	sp "github.com/963sup/line-mini-app-bot/internal/business/space"
	spaceDB "github.com/963sup/line-mini-app-bot/internal/business/space/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type membership struct{}

func (membership) IsMember(_ context.Context, _, actor string) (bool, error) {
	return actor != "outsider", nil
}

type sourceAccess struct {
	root   *db.DocumentRef
	denied bool
}

func (s *sourceAccess) Authorize(ctx context.Context, actor, id, spaceID string) (p.Source, error) {
	if s.denied || actor == "outsider" {
		return p.Source{}, p.ErrNotFound
	}
	snap, err := s.root.Collection("repositories").Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return p.Source{}, p.ErrNotFound
	}
	if err != nil {
		return p.Source{}, err
	}
	d := snap.Data()
	group, _ := d["groupId"].(string)
	if d["id"] != id || d["status"] != "ACTIVE" || spaceID != "" && spaceID != group {
		return p.Source{}, p.ErrNotFound
	}
	return p.Source{ID: id, SpaceID: group}, nil
}

type auth struct{}

func (auth) Authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	a := r.Header.Get("X-Test-Actor")
	if a == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return "", "", false
	}
	return a, "", true
}

func TestSourceHydrationFailureDoesNotHideSpaceMetadata(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires an isolated FIRESTORE_EMULATOR_HOST")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "project-decouple-local")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	channel := "hydration-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	root := client.Collection("channels").Doc(channel)
	gid, rid, iid := "C"+strings.Repeat("a", 32), strings.Repeat("b", 32), strings.Repeat("c", 32)
	spaces := &spaceDB.Store{Client: client, ChannelID: channel}
	if err = spaces.Observe(ctx, sp.Observation{GroupID: gid, Status: "ACTIVE", EventID: "join", At: time.Now().UTC(), Users: []string{"owner"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = root.Collection("repositories").Doc(rid).Set(ctx, map[string]any{"id": rid, "groupId": gid, "status": "ACTIVE", "version": int64(1)}); err != nil {
		t.Fatal(err)
	}
	issue := root.Collection("repositories").Doc(rid).Collection("issues").Doc(iid)
	issueData := map[string]any{"schemaVersion": int64(1), "id": iid, "repositoryId": rid, "title": "source", "state": "open"}
	if _, err = issue.Set(ctx, issueData); err != nil {
		t.Fatal(err)
	}
	store := &projectDB.Store{Client: client, ChannelID: channel}
	svc := &p.Service{Spaces: &sp.Service{Store: spaces, Membership: membership{}}, Sources: &sourceAccess{root: root}, Store: store}
	created, err := svc.Create(ctx, "owner", gid, "source-failure-key-001", p.Content{Title: "owned metadata", SourceRepositoryID: rid})
	if err != nil {
		t.Fatal(err)
	}
	added, err := svc.ChangeItem(ctx, "owner", gid, created.ID, iid, true, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = issue.Delete(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(ctx, "owner", gid, created.ID)
	if err != nil || got.Title != created.Title || got.SourceAvailable || got.IssueCount != 0 || got.CompletedIssueCount != 0 {
		t.Fatalf("source read failure must preserve metadata without partial counts: %+v %v", got, err)
	}
	page, err := svc.List(ctx, "owner", gid, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].SourceAvailable {
		t.Fatalf("source failure hid Space list: %+v %v", page, err)
	}
	updated, err := svc.Update(ctx, "owner", gid, created.ID, p.Update{Content: p.Content{Title: "updated metadata"}, ExpectedVersion: added.Version})
	if err != nil || updated.Version != added.Version+1 || updated.SourceAvailable {
		t.Fatalf("committed own metadata update failed hydration: %+v %v", updated, err)
	}
	if _, err = svc.Get(ctx, "owner", rid, created.ID); !errors.Is(err, p.ErrInput) {
		t.Fatalf("legacy source error changed: %v", err)
	}
	if _, err = svc.ListItems(ctx, "owner", gid, created.ID, ""); !errors.Is(err, p.ErrInput) {
		t.Fatalf("source item read must still fail: %v", err)
	}
	value := updated
	err = store.Progress(ctx, p.Scope{SpaceID: gid, Actor: "owner"}, &value)
	if !errors.Is(err, p.ErrSourceUnavailable) || !errors.Is(err, p.ErrUnavailable) || errors.Is(err, p.ErrInput) {
		t.Fatalf("source classification must expose only its domain categories: %v", err)
	}
	if _, err = issue.Set(ctx, issueData); err != nil {
		t.Fatal(err)
	}
	items, err := root.Collection("spaces").Doc(gid).Collection("projects").Doc(created.ID).Collection("items").Documents(ctx).GetAll()
	if err != nil || len(items) != 1 {
		t.Fatalf("items fixture: %v %v", items, err)
	}
	if _, err = items[0].Ref.Update(ctx, []db.Update{{Path: "schemaVersion", Value: int64(0)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Get(ctx, "owner", gid, created.ID); !errors.Is(err, p.ErrUnavailable) || errors.Is(err, p.ErrSourceUnavailable) {
		t.Fatalf("own item corruption must fail closed: %v", err)
	}
}

func TestSpaceProjectStorageRightsAndLegacyAlias(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires an isolated FIRESTORE_EMULATOR_HOST")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "project-decouple-local")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	channel := "project-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	root := client.Collection("channels").Doc(channel)
	gid := "C" + strings.Repeat("a", 32)
	rid := strings.Repeat("b", 32)
	iid := strings.Repeat("c", 32)
	ss := &spaceDB.Store{Client: client, ChannelID: channel}
	if err = ss.Observe(ctx, sp.Observation{GroupID: gid, Status: "ACTIVE", EventID: "join", At: time.Now().UTC(), Users: []string{"owner", "member"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = root.Collection("repositories").Doc(rid).Set(ctx, map[string]any{"id": rid, "groupId": gid, "status": "ACTIVE", "version": int64(1)}); err != nil {
		t.Fatal(err)
	}
	issue := root.Collection("repositories").Doc(rid).Collection("issues").Doc(iid)
	sourceData := map[string]any{"schemaVersion": int64(1), "id": iid, "repositoryId": rid, "title": "Source", "state": "open"}
	if _, err = issue.Set(ctx, sourceData); err != nil {
		t.Fatal(err)
	}
	source := &sourceAccess{root: root}
	store := &projectDB.Store{Client: client, ChannelID: channel}
	svc := &p.Service{Spaces: &sp.Service{Store: ss, Membership: membership{}}, Sources: source, Store: store}
	content := p.Content{Title: "Space project", SourceRepositoryID: rid}
	key := "stable-create-key-0001"
	if _, err = svc.Create(ctx, "outsider", gid, key, content); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("Space outsider create: %v", err)
	}
	created, err := svc.Create(ctx, "owner", gid, key, content)
	if err != nil {
		t.Fatal(err)
	}
	if created.SpaceID != gid || created.SourceRepositoryID != rid || created.Version != 1 {
		t.Fatalf("Project owner is Space; source is a reference: %+v", created)
	}
	if _, err = root.Collection("repositories").Doc(rid).Collection("projects").Doc(created.ID).Get(ctx); status.Code(err) != codes.NotFound {
		t.Fatalf("new Project must not write legacy container: %v", err)
	}
	if _, err = root.Collection("spaces").Doc(gid).Collection("projects").Doc(created.ID).Get(ctx); err != nil {
		t.Fatal(err)
	}
	repoStore := &repoDB.Store{Client: client, ChannelID: channel}
	counts, err := repoStore.Statistics(ctx, repository.Repository{ID: rid, GroupID: gid, Status: "ACTIVE"})
	if err != nil || counts.Projects != 1 {
		t.Fatalf("statistics must count new canonical project without legacy copy: %+v %v", counts, err)
	}
	// Old and new endpoints share the exact creation receipt and aggregate identity.
	replay, err := svc.Create(ctx, "owner", rid, key, p.Content{Title: content.Title})
	if err != nil || replay.ID != created.ID {
		t.Fatalf("legacy receipt replay: %+v %v", replay, err)
	}
	var wg sync.WaitGroup
	results := make(chan p.Project, 3)
	failures := make(chan error, 3)
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := svc.Create(ctx, "owner", gid, key, content)
			if e != nil {
				failures <- e
			} else {
				results <- v
			}
		}()
	}
	wg.Wait()
	close(results)
	close(failures)
	for e := range failures {
		t.Fatal(e)
	}
	for v := range results {
		if v.ID != created.ID {
			t.Fatal("same creation key produced another identity")
		}
	}
	if _, err = svc.Create(ctx, "owner", gid, key, p.Content{Title: "different", SourceRepositoryID: rid}); !errors.Is(err, p.ErrCreationConflict) {
		t.Fatalf("key conflict: %v", err)
	}
	if _, err = svc.Update(ctx, "member", gid, created.ID, p.Update{Content: p.Content{Title: "forbidden"}, ExpectedVersion: 1}); !errors.Is(err, p.ErrForbidden) {
		t.Fatalf("membership does not grant Project management: %v", err)
	}
	added, err := svc.ChangeItem(ctx, "owner", gid, created.ID, iid, true, 1)
	if err != nil {
		t.Fatal(err)
	}
	if added.Version != 2 || added.IssueCount != 1 || !added.SourceAvailable {
		t.Fatalf("mutation result must preserve live source progress: %+v", added)
	}
	if _, err = svc.ChangeItem(ctx, "owner", rid, created.ID, iid, true, 1); !errors.Is(err, p.ErrConflict) {
		t.Fatalf("legacy CAS must share new version: %v", err)
	}
	page, err := svc.ListItems(ctx, "member", gid, created.ID, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].IssueID != iid || page.Items[0].ID == iid {
		t.Fatalf("independent item identity/source read: %+v %v", page, err)
	}
	legacy, err := svc.Get(ctx, "member", rid, created.ID)
	if err != nil || legacy.ID != created.ID || legacy.Version != 2 || legacy.IssueCount != 1 {
		t.Fatalf("legacy read alias: %+v %v", legacy, err)
	}
	mux := http.NewServeMux()
	projectHTTP.Register(mux, svc, auth{})
	call := func(method, path, actor, payload string) (int, map[string]any) {
		req := httptest.NewRequest(method, path, strings.NewReader(payload))
		if actor != "" {
			req.Header.Set("X-Test-Actor", actor)
		}
		if payload != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		var result map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &result)
		return rec.Code, result
	}
	canonicalPath := "/api/v1/spaces/" + gid + "/projects/" + created.ID
	legacyPath := "/api/v1/repositories/" + rid + "/projects/" + created.ID
	for _, path := range []string{canonicalPath, legacyPath} {
		if code, body := call(http.MethodGet, path, "member", ""); code != 200 || body["id"] != created.ID || body["version"] != float64(2) || body["issueCount"] != float64(1) {
			t.Fatalf("canonical/legacy HTTP read alias %s: %d %+v", path, code, body)
		}
		if code, _ := call(http.MethodGet, path, "", ""); code != 401 {
			t.Fatalf("missing HTTP identity %s: %d", path, code)
		}
		if code, _ := call(http.MethodGet, path, "outsider", ""); code != 404 {
			t.Fatalf("outsider HTTP must deny %s: %d", path, code)
		}
	}
	if code, _ := call(http.MethodPatch, canonicalPath, "member", `{"title":"forbidden","body":"","expectedVersion":2}`); code != 403 {
		t.Fatalf("HTTP project creator rule: %d", code)
	}
	if code, _ := call(http.MethodPatch, canonicalPath, "owner", `{"title":"stale","body":"","expectedVersion":1}`); code != 409 {
		t.Fatalf("HTTP shared version CAS: %d", code)
	}
	source.denied = true
	metadata, err := svc.Get(ctx, "owner", gid, created.ID)
	if err != nil || metadata.Title != content.Title || metadata.SourceAvailable {
		t.Fatalf("own metadata must survive revoked source rights: %+v %v", metadata, err)
	}
	if _, err = svc.ListItems(ctx, "owner", gid, created.ID, ""); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("source revoke must hide source issues: %v", err)
	}
	if _, err = svc.ChangeItem(ctx, "owner", gid, created.ID, iid, false, 2); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("source revoke must stop source-linked actions: %v", err)
	}
	if _, err = svc.Get(ctx, "owner", rid, created.ID); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("legacy source-scoped read retains denial: %v", err)
	}
	updated, err := svc.Update(ctx, "owner", gid, created.ID, p.Update{Content: p.Content{Title: "Own edit"}, ExpectedVersion: 2})
	if err != nil || updated.Version != 3 {
		t.Fatalf("own metadata edit independent of source: %+v %v", updated, err)
	}
	snapshot, err := issue.Get(ctx)
	if err != nil || !reflect.DeepEqual(snapshot.Data(), sourceData) {
		t.Fatalf("Project operation mutated source Issue: %+v %v", snapshot, err)
	}
	marker := root.Collection("projectStorage").Doc(rid)
	if _, err = marker.Update(ctx, []db.Update{{Path: "locked", Value: true}}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Get(ctx, "owner", gid, created.ID); !errors.Is(err, p.ErrUnavailable) {
		t.Fatalf("canonical read cannot expose partial migration copies under lock: %v", err)
	}
	if _, err = svc.List(ctx, "owner", gid, ""); !errors.Is(err, p.ErrUnavailable) {
		t.Fatalf("canonical list cannot expose partial migration copies under lock: %v", err)
	}
	if _, err = svc.Update(ctx, "owner", gid, created.ID, p.Update{Content: p.Content{Title: "blocked by migration"}, ExpectedVersion: 3}); !errors.Is(err, p.ErrUnavailable) {
		t.Fatalf("migration lock must block canonical writers: %v", err)
	}
}

func TestLegacyWriteChecksCurrentSpaceAndStatisticsRouting(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires an isolated FIRESTORE_EMULATOR_HOST")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	client, err := db.NewClient(ctx, "project-decouple-local")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	channel := "legacy-space-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	root := client.Collection("channels").Doc(channel)
	gid := "C" + strings.Repeat("d", 32)
	rid := strings.Repeat("e", 32)
	ss := &spaceDB.Store{Client: client, ChannelID: channel}
	if err = ss.Observe(ctx, sp.Observation{GroupID: gid, Status: "ACTIVE", EventID: "join", At: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	repoRef := root.Collection("repositories").Doc(rid)
	if _, err = repoRef.Set(ctx, map[string]any{"id": rid, "groupId": gid, "status": "ACTIVE", "version": int64(1)}); err != nil {
		t.Fatal(err)
	}
	store := &projectDB.Store{Client: client, ChannelID: channel}
	scope := p.Scope{SpaceID: gid, SourceRepositoryID: rid, Actor: "owner", Legacy: true}
	if _, err = root.Collection("projectStorage").Doc(rid).Set(ctx, map[string]any{"spaceId": gid, "location": "legacy", "locked": false}); err != nil {
		t.Fatal(err)
	}
	legacy, err := store.Create(ctx, scope, "owner", "legacy-create-key-0001", p.Content{Title: "Legacy", SourceRepositoryID: rid})
	if err != nil {
		t.Fatal(err)
	}
	stats := &repoDB.Store{Client: client, ChannelID: channel}
	r := repository.Repository{ID: rid, GroupID: gid, Status: "ACTIVE"}
	before, err := stats.Statistics(ctx, r)
	if err != nil || before.Projects != 1 {
		t.Fatalf("legacy statistics count: %+v %v", before, err)
	}
	old, err := repoRef.Collection("projects").Doc(legacy.ID).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	data := old.Data()
	data["spaceId"], data["sourceRepositoryId"] = gid, rid
	if _, err = root.Collection("spaces").Doc(gid).Collection("projects").Doc(legacy.ID).Set(ctx, data); err != nil {
		t.Fatal(err)
	}
	if _, err = root.Collection("projectStorage").Doc(rid).Set(ctx, map[string]any{"spaceId": gid, "location": "space", "locked": false}); err != nil {
		t.Fatal(err)
	}
	after, err := stats.Statistics(ctx, r)
	if err != nil || after.Projects != 1 {
		t.Fatalf("migration marker must select canonical once, not double count: %+v %v", after, err)
	}
	// A clean rollback changes only the routing pointer; canonical APIs must stay useful.
	marker := root.Collection("projectStorage").Doc(rid)
	if _, err = marker.Update(ctx, []db.Update{{Path: "location", Value: "legacy"}}); err != nil {
		t.Fatal(err)
	}
	source := &sourceAccess{root: root}
	svc := &p.Service{Spaces: &sp.Service{Store: ss, Membership: membership{}}, Sources: source, Store: store}
	visible, err := svc.List(ctx, "owner", gid, "")
	if err != nil || len(visible.Items) != 1 || visible.Items[0].ID != legacy.ID {
		t.Fatalf("canonical list must resolve retained legacy records after rollback: %+v %v", visible, err)
	}
	replay, err := svc.Create(ctx, "owner", rid, "legacy-create-key-0001", p.Content{Title: "Legacy"})
	if err != nil || replay.ID != legacy.ID || replay.Version != legacy.Version {
		t.Fatalf("rollback must retain original receipt identity/version: %+v %v", replay, err)
	}
	if _, err = repoRef.Update(ctx, []db.Update{{Path: "status", Value: "INACTIVE"}}); err != nil {
		t.Fatal(err)
	}
	own, err := svc.Get(ctx, "owner", gid, legacy.ID)
	if err != nil || own.ID != legacy.ID || own.SourceAvailable {
		t.Fatalf("canonical metadata on legacy layout remains Space-authorized with revoked source: %+v %v", own, err)
	}
	if _, err = svc.ListItems(ctx, "owner", gid, legacy.ID, ""); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("rollback must not expand revoked source read: %v", err)
	}
	edited, err := svc.Update(ctx, "owner", gid, legacy.ID, p.Update{Content: p.Content{Title: "Owner edit after rollback"}, ExpectedVersion: 1})
	if err != nil || edited.Version != 2 {
		t.Fatalf("canonical own metadata update cannot treat legacy Repository as authorization container: %+v %v", edited, err)
	}
	if _, err = svc.Get(ctx, "owner", rid, legacy.ID); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("legacy endpoint preserves source rejection: %v", err)
	}
	// Simulate a new validated copy after the legacy write, then switch once again.
	old, err = repoRef.Collection("projects").Doc(legacy.ID).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	data = old.Data()
	data["spaceId"], data["sourceRepositoryId"] = gid, rid
	if _, err = root.Collection("spaces").Doc(gid).Collection("projects").Doc(legacy.ID).Set(ctx, data); err != nil {
		t.Fatal(err)
	}
	if _, err = repoRef.Update(ctx, []db.Update{{Path: "status", Value: "ACTIVE"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = marker.Update(ctx, []db.Update{{Path: "location", Value: "space"}}); err != nil {
		t.Fatal(err)
	}
	canonicalScope := scope
	canonicalScope.Legacy = false
	if _, err = store.Create(ctx, canonicalScope, "owner", "canonical-create-key-0001", p.Content{Title: "Canonical", SourceRepositoryID: rid}); err != nil {
		t.Fatal(err)
	}
	after, err = stats.Statistics(ctx, r)
	if err != nil || after.Projects != 2 {
		t.Fatalf("canonical creates must remain in source-derived statistics: %+v %v", after, err)
	}
	// A request may have authorized before the owner Space deactivated.
	if _, err = root.Collection("spaces").Doc(gid).Update(ctx, []db.Update{{Path: "status", Value: "INACTIVE"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Update(ctx, scope, "owner", legacy.ID, p.Update{Content: p.Content{Title: "must not write"}, ExpectedVersion: 1}); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("legacy writer must recheck owner Space in its transaction: %v", err)
	}
	// The absent-marker legacy parent needs the same owner check, not Repository status alone.
	if _, err = root.Collection("projectStorage").Doc(rid).Delete(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Update(ctx, scope, "owner", legacy.ID, p.Update{Content: p.Content{Title: "must not write"}, ExpectedVersion: 1}); !errors.Is(err, p.ErrNotFound) {
		t.Fatalf("legacy-layout writer must reject inactive Space despite active Repository: %v", err)
	}
}
