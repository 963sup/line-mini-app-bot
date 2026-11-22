package firestore_test

import (
	projectgroup "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/group"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

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

	spaceDB "github.com/963sup/line-mini-app-bot/internal/business/group/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/business/group/application/query"
	groupevent "github.com/963sup/line-mini-app-bot/internal/business/group/domain/event"

	projectDB "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	projectHTTP "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/http"

	repoDB "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type membership struct{}

func (membership) IsMember(_ context.Context, _, actor string) (bool, error) {
	return actor != "outsider", nil
}

type sourceAccess struct {
	root               *db.DocumentRef
	denied             bool
	deniedRepositoryID string
}

func (s *sourceAccess) Authorize(ctx context.Context, actor, id, expectedOwnerID string) (projectDomainEntity.Source, error) {
	if s.denied || id == s.deniedRepositoryID || actor == "outsider" {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrNotFound
	}
	snap, err := s.root.Collection("repositories").Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrNotFound
	}
	if err != nil {
		return projectDomainEntity.Source{}, err
	}
	d := snap.Data()
	owner, _ := d["ownerId"].(string)
	if d["id"] != id || d["status"] != "ACTIVE" || expectedOwnerID != "" && expectedOwnerID != owner {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrNotFound
	}
	return projectDomainEntity.Source{ID: id, OwnerID: owner}, nil
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

func TestSourceHydrationFailurePreservesProjectMetadata(t *testing.T) {
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
	groups := &spaceDB.Store{Client: client, ChannelID: channel}
	if err = groups.Observe(ctx, groupevent.Observation{GroupID: gid, Status: "ACTIVE", EventID: "join", At: time.Now().UTC(), Users: []string{"owner"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = root.Collection("repositories").Doc(rid).Set(ctx, map[string]any{"id": rid, "ownerId": gid, "status": "ACTIVE", "version": int64(1)}); err != nil {
		t.Fatal(err)
	}
	issue := root.Collection("repositories").Doc(rid).Collection("issues").Doc(iid)
	issueData := map[string]any{"schemaVersion": int64(1), "id": iid, "repositoryId": rid, "title": "source", "state": "open"}
	if _, err = issue.Set(ctx, issueData); err != nil {
		t.Fatal(err)
	}
	store := &projectDB.Store{Client: client, ChannelID: channel}
	svc := &projectApplication.Service{Groups: projectgroup.Access{Groups: &groupquery.Service{Store: groups, Membership: membership{}}}, Sources: &sourceAccess{root: root}, Store: store}
	created, err := svc.Create(ctx, "owner", gid, "source-failure-key-001", projectDomainEntity.Content{Title: "owned metadata", SourceRepositoryID: rid})
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
		t.Fatalf("source failure hid Group list: %+v %v", page, err)
	}
	updated, err := svc.Update(ctx, "owner", gid, created.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: "updated metadata"}, ExpectedVersion: added.Version})
	if err != nil || updated.Version != added.Version+1 || updated.SourceAvailable {
		t.Fatalf("committed own metadata update failed hydration: %+v %v", updated, err)
	}
	if _, err = svc.Get(ctx, "owner", rid, created.ID); !errors.Is(err, projectDomainEntity.ErrInput) {
		t.Fatalf("legacy source error changed: %v", err)
	}
	if _, err = svc.ListItems(ctx, "owner", gid, created.ID, ""); !errors.Is(err, projectDomainEntity.ErrInput) {
		t.Fatalf("source item read must still fail: %v", err)
	}
	value := updated
	err = store.Progress(ctx, projectApplicationPort.Scope{OwnerID: gid, Actor: "owner"}, &value)
	if !errors.Is(err, projectDomainEntity.ErrSourceUnavailable) || errors.Is(err, projectDomainEntity.ErrInput) {
		t.Fatalf("source classification must expose only its domain categories: %v", err)
	}
	if _, err = issue.Set(ctx, issueData); err != nil {
		t.Fatal(err)
	}
	items, err := root.Collection("groups").Doc(gid).Collection("projects").Doc(created.ID).Collection("items").Documents(ctx).GetAll()
	if err != nil || len(items) != 1 {
		t.Fatalf("items fixture: %v %v", items, err)
	}
	if _, err = items[0].Ref.Update(ctx, []db.Update{{Path: "schemaVersion", Value: int64(0)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Get(ctx, "owner", gid, created.ID); !errors.Is(err, projectDomainEntity.ErrUnavailable) || errors.Is(err, projectDomainEntity.ErrSourceUnavailable) {
		t.Fatalf("own item corruption must fail closed: %v", err)
	}
}

func TestOwnerProjectStorageRightsAndLegacyAlias(t *testing.T) {
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
	if err = ss.Observe(ctx, groupevent.Observation{GroupID: gid, Status: "ACTIVE", EventID: "join", At: time.Now().UTC(), Users: []string{"owner", "member"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = root.Collection("repositories").Doc(rid).Set(ctx, map[string]any{"id": rid, "ownerId": gid, "status": "ACTIVE", "version": int64(1)}); err != nil {
		t.Fatal(err)
	}
	issue := root.Collection("repositories").Doc(rid).Collection("issues").Doc(iid)
	sourceData := map[string]any{"schemaVersion": int64(1), "id": iid, "repositoryId": rid, "title": "Source", "state": "open"}
	if _, err = issue.Set(ctx, sourceData); err != nil {
		t.Fatal(err)
	}
	source := &sourceAccess{root: root}
	store := &projectDB.Store{Client: client, ChannelID: channel}
	svc := &projectApplication.Service{Groups: projectgroup.Access{Groups: &groupquery.Service{Store: ss, Membership: membership{}}}, Sources: source, Store: store}
	content := projectDomainEntity.Content{Title: "Group project", SourceRepositoryID: rid}
	key := "stable-create-key-0001"
	if _, err = svc.Create(ctx, "outsider", gid, key, content); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatalf("Group outsider create: %v", err)
	}
	created, err := svc.Create(ctx, "owner", gid, key, content)
	if err != nil {
		t.Fatal(err)
	}
	if created.OwnerID != gid || created.DefaultRepositoryID != rid || created.Version != 1 {
		t.Fatalf("Project owner is Group; source is a reference: %+v", created)
	}
	if _, err = root.Collection("repositories").Doc(rid).Collection("projects").Doc(created.ID).Get(ctx); status.Code(err) != codes.NotFound {
		t.Fatalf("new Project must not write legacy container: %v", err)
	}
	if _, err = root.Collection("groups").Doc(gid).Collection("projects").Doc(created.ID).Get(ctx); err != nil {
		t.Fatal(err)
	}
	repoStore := &repoDB.Store{Client: client, ChannelID: channel}
	counts, err := repoStore.Statistics(ctx, repositoryDomainEntity.Repository{ID: rid, OwnerID: gid, Status: "ACTIVE"})
	if err != nil || counts.Projects != 1 {
		t.Fatalf("statistics must count new canonical project without legacy copy: %+v %v", counts, err)
	}
	// Old and new endpoints share the exact creation receipt and aggregate identity.
	replay, err := svc.Create(ctx, "owner", rid, key, projectDomainEntity.Content{Title: content.Title})
	if err != nil || replay.ID != created.ID {
		t.Fatalf("legacy receipt replay: %+v %v", replay, err)
	}
	var wg sync.WaitGroup
	results := make(chan projectDomainEntity.Project, 3)
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
	if _, err = svc.Create(ctx, "owner", gid, key, projectDomainEntity.Content{Title: "different", SourceRepositoryID: rid}); !errors.Is(err, projectDomainEntity.ErrCreationConflict) {
		t.Fatalf("key conflict: %v", err)
	}
	if _, err = svc.Update(ctx, "member", gid, created.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: "forbidden"}, ExpectedVersion: 1}); !errors.Is(err, projectDomainEntity.ErrForbidden) {
		t.Fatalf("membership does not grant Project management: %v", err)
	}
	added, err := svc.ChangeItem(ctx, "owner", gid, created.ID, iid, true, 1)
	if err != nil {
		t.Fatal(err)
	}
	if added.Version != 2 || added.IssueCount != 1 || !added.SourceAvailable {
		t.Fatalf("mutation result must preserve live source progress: %+v", added)
	}
	if _, err = svc.ChangeItem(ctx, "owner", rid, created.ID, iid, true, 1); !errors.Is(err, projectDomainEntity.ErrConflict) {
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
	canonicalPath := "/api/v1/owners/" + gid + "/projects/" + created.ID
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
	if _, err = svc.ListItems(ctx, "owner", gid, created.ID, ""); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatalf("source revoke must hide source issues: %v", err)
	}
	if _, err = svc.ChangeItem(ctx, "owner", gid, created.ID, iid, false, 2); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatalf("source revoke must stop source-linked actions: %v", err)
	}
	if _, err = svc.Get(ctx, "owner", rid, created.ID); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatalf("legacy source-scoped read retains denial: %v", err)
	}
	updated, err := svc.Update(ctx, "owner", gid, created.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: "Own edit"}, ExpectedVersion: 2})
	if err != nil || updated.Version != 3 {
		t.Fatalf("own metadata edit independent of source: %+v %v", updated, err)
	}
	snapshot, err := issue.Get(ctx)
	if err != nil || !reflect.DeepEqual(snapshot.Data(), sourceData) {
		t.Fatalf("Project operation mutated source Issue: %+v %v", snapshot, err)
	}
	// Unrelated or obsolete routing metadata cannot redirect the canonical Group.
	if _, err = root.Collection("projectStorage").Doc(rid).Set(ctx, map[string]any{"location": "legacy", "locked": true}); err != nil {
		t.Fatal(err)
	}
	own, err := svc.Get(ctx, "owner", gid, created.ID)
	if err != nil || own.Title != updated.Title {
		t.Fatal(own, err)
	}
	if _, err = root.Collection("groups").Doc(gid).Update(ctx, []db.Update{{Path: "status", Value: "INACTIVE"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.RunUpdate(ctx, projectApplicationPort.Scope{OwnerID: gid, Actor: "owner"}, created.ID, func(projectApplicationPort.UpdateWork) error {
		t.Fatal("inactive project scope reached update")
		return nil
	}); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatal("inactive Group allowed write", err)
	}
}

func TestPersonalRepositoryAndProject(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST")
	}
	ctx := t.Context()
	client, err := db.NewClient(ctx, "personal-project-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	channel := "personal-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	root := client.Collection("channels").Doc(channel)
	owner, other := "U"+strings.Repeat("a", 32), "U"+strings.Repeat("b", 32)
	repos := &repoDB.Store{Client: client, ChannelID: channel}
	seen := repositoryDomainEntity.Observation{OwnerID: owner, Name: "個人儲存庫", Status: "ACTIVE", EventID: "follow", At: time.Now(), Users: []string{owner}}
	for range 2 {
		if err := repos.Observe(ctx, seen); err != nil {
			t.Fatal(err)
		}
	}
	catalog := repositoryApplication.Service{Store: repos}
	page, err := catalog.List(ctx, owner, "", "")
	if err != nil || len(page.Items) != 1 || page.Items[0].OwnerID != owner {
		t.Fatalf("personal list: %+v %v", page, err)
	}
	rid := page.Items[0].ID
	if _, err = catalog.Get(ctx, other, rid); !errors.Is(err, repositoryDomainEntity.ErrNotFound) {
		t.Fatal("foreign repository read", err)
	}
	access := &sourceAccess{root: root}
	svc := projectApplication.Service{Groups: projectgroup.Access{Groups: &groupquery.Service{}}, Sources: access, Store: &projectDB.Store{Client: client, ChannelID: channel}}
	value, err := svc.Create(ctx, owner, owner, "personal-project-0001", projectDomainEntity.Content{Title: "工程", SourceRepositoryID: rid})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := root.Collection("users").Doc(owner).Collection("projects").Doc(value.ID).Get(ctx)
	if err != nil || doc.Data()["ownerId"] != owner {
		t.Fatal("personal project storage", err)
	}
	if _, err = svc.Get(ctx, other, owner, value.ID); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatal("foreign project read", err)
	}
	group := "C" + strings.Repeat("c", 32)
	groupRepo := strings.Repeat("d", 32)
	groupDoc := root.Collection("repositories").Doc(groupRepo)
	if _, err = groupDoc.Set(ctx, map[string]any{"id": groupRepo, "ownerId": group, "status": "ACTIVE"}); err != nil {
		t.Fatal(err)
	}
	issue := strings.Repeat("e", 32)
	if _, err = groupDoc.Collection("issues").Doc(issue).Set(ctx, map[string]any{"schemaVersion": 1, "id": issue, "repositoryId": groupRepo, "title": "Group issue", "state": "open"}); err != nil {
		t.Fatal(err)
	}
	value, err = svc.Update(ctx, owner, owner, value.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: value.Title}, ExpectedVersion: value.Version, LinkRepositoryID: groupRepo})
	if err != nil {
		t.Fatal(err)
	}
	value, err = svc.AddReference(ctx, owner, owner, value.ID, projectDomainEntity.Reference{SourceRepositoryID: groupRepo, SourceType: "issue", SourceID: issue}, value.Version)
	if err != nil {
		t.Fatal(err)
	}
	before, err := svc.ListReferences(ctx, owner, owner, value.ID, "")
	if err != nil || len(before.Items) != 1 || before.Items[0].SourceTitle != "Group issue" {
		t.Fatal(before, err)
	}
	access.deniedRepositoryID = groupRepo
	after, err := svc.ListReferences(ctx, owner, owner, value.ID, "")
	if err != nil || len(after.Items) != 1 || !after.Items[0].SourceUnavailable || after.Items[0].SourceTitle != "" {
		t.Fatal(after, err)
	}
	current, err := svc.Get(ctx, owner, owner, value.ID)
	if err != nil || current.OwnerID != owner {
		t.Fatal(current, err)
	}
	repo, err := groupDoc.Get(ctx)
	if err != nil || repo.Data()["ownerId"] != group {
		t.Fatal("group repository owner changed", err)
	}
}
