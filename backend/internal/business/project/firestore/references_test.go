package firestore_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	p "github.com/963sup/line-mini-app-bot/internal/business/project"
	projectDB "github.com/963sup/line-mini-app-bot/internal/business/project/firestore"
	projectHTTP "github.com/963sup/line-mini-app-bot/internal/business/project/http"
	sp "github.com/963sup/line-mini-app-bot/internal/business/space"
	spaceDB "github.com/963sup/line-mini-app-bot/internal/business/space/firestore"
)

func TestEmulatorProjectTypedReferencesLatestAndRevokedSources(t *testing.T) {
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("requires loopback emulator")
	}
	hostname, _, e := net.SplitHostPort(host)
	if e != nil || (hostname != "localhost" && !net.ParseIP(hostname).IsLoopback()) {
		t.Fatal("loopback only")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var random [8]byte
	if _, e = rand.Read(random[:]); e != nil {
		t.Fatal(e)
	}
	client, e := db.NewClient(ctx, "project-refs-"+hex.EncodeToString(random[:]))
	if e != nil {
		t.Fatal(e)
	}
	defer func() { _ = client.Close() }()
	root := client.Collection("channels").Doc("isolated-refs")
	gid, rid := "C"+strings.Repeat("a", 32), strings.Repeat("b", 32)
	spaces := &spaceDB.Store{Client: client, ChannelID: root.ID}
	if e = spaces.Observe(ctx, sp.Observation{GroupID: gid, Status: "ACTIVE", EventID: "synthetic-join", At: time.Now().UTC(), Users: []string{"owner"}}); e != nil {
		t.Fatal(e)
	}
	repository := root.Collection("repositories").Doc(rid)
	if _, e = repository.Set(ctx, map[string]any{"id": rid, "groupId": gid, "status": "ACTIVE", "version": int64(1)}); e != nil {
		t.Fatal(e)
	}
	oid, sid, iid, did := strings.Repeat("c", 32), strings.Repeat("d", 32), strings.Repeat("e", 32), strings.Repeat("f", 32)
	object := repository.Collection("objects").Doc(oid)
	sub := object.Collection("subobjects").Doc(sid)
	issue := repository.Collection("issues").Doc(iid)
	discussion := repository.Collection("discussions").Doc(did)
	data := []struct {
		ref    *db.DocumentRef
		fields map[string]any
	}{{object, map[string]any{"schemaVersion": 1, "id": oid, "repositoryId": rid, "name": "Fiber", "type": "work", "quantity": "140", "unit": "M", "completedQuantity": "40", "status": "in_progress"}}, {sub, map[string]any{"schemaVersion": 1, "id": sid, "repositoryId": rid, "parentObjectId": oid, "name": "Splice", "type": "work", "quantity": "11", "unit": "SET", "completedQuantity": "0", "status": "not_started"}}, {issue, map[string]any{"schemaVersion": 1, "id": iid, "repositoryId": rid, "title": "Question", "state": "completed"}}, {discussion, map[string]any{"schemaVersion": 1, "id": did, "repositoryId": rid, "title": "Discussion"}}}
	for _, source := range data {
		if _, e = source.ref.Set(ctx, source.fields); e != nil {
			t.Fatal(e)
		}
	}
	access := &sourceAccess{root: root}
	store := &projectDB.Store{Client: client, ChannelID: root.ID}
	svc := &p.Service{Spaces: &sp.Service{Store: spaces, Membership: membership{}}, Sources: access, Store: store}
	project, e := svc.Create(ctx, "owner", gid, "typed-reference-project", p.Content{Title: "Existing project", SourceRepositoryID: rid})
	if e != nil {
		t.Fatal(e)
	}
	project, e = svc.ChangeItem(ctx, "owner", gid, project.ID, iid, true, project.Version)
	if e != nil {
		t.Fatal(e)
	}
	oldItems, e := svc.ListItems(ctx, "owner", gid, project.ID, "")
	if e != nil {
		t.Fatal(e)
	}
	oldIssueID := oldItems.Items[0].ID
	refs := []p.Reference{{SourceType: "issue", SourceID: iid}, {SourceType: "object", SourceID: oid}, {SourceType: "subobject", SourceID: sid, ParentObjectID: oid}, {SourceType: "discussion", SourceID: did}}
	for _, ref := range refs {
		priorVersion := project.Version
		project, e = svc.AddReference(ctx, "owner", gid, project.ID, ref, project.Version)
		if e != nil {
			t.Fatal(ref, e)
		}
		replay, e := svc.AddReference(ctx, "owner", gid, project.ID, ref, priorVersion)
		if e != nil || replay.Version != project.Version {
			t.Fatal("unknown retry", replay, e)
		}
	}
	if project.IssueCount != 1 || project.CompletedIssueCount != 1 {
		t.Fatal("mixed counts", project)
	}
	page, e := svc.ListReferences(ctx, "owner", gid, project.ID, "")
	if e != nil || len(page.Items) != 4 {
		t.Fatal(page, e)
	}
	ids := map[string]string{}
	for _, ref := range page.Items {
		ids[ref.SourceType] = ref.ID
		if ref.SourceUnavailable || ref.SourceTitle == "" {
			t.Fatal(ref)
		}
	}
	if ids["issue"] != oldIssueID {
		t.Fatal("duplicate legacy issue")
	}
	raw, e := root.Collection("spaces").Doc(gid).Collection("projects").Doc(project.ID).Collection("items").Documents(ctx).GetAll()
	if e != nil || len(raw) != 4 {
		t.Fatal(e)
	}
	for _, snap := range raw {
		for _, forbidden := range []string{"title", "name", "sourceTitle", "quantity", "completedQuantity"} {
			if _, ok := snap.Data()[forbidden]; ok {
				t.Fatal("snapshot stored", forbidden)
			}
		}
	}
	for _, source := range data {
		snap, e := source.ref.Get(ctx)
		if e != nil {
			t.Fatal(e)
		}
		encoded, _ := json.Marshal(snap.Data())
		wanted, _ := json.Marshal(source.fields)
		if string(encoded) != string(wanted) { // Integer encodings in JSON normalize Firestore int64 and fixture int.
			t.Fatal("source changed", string(encoded), string(wanted))
		}
	}
	before, e := root.Collection("spaces").Doc(gid).Collection("projects").Doc(project.ID).Get(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = object.Update(ctx, []db.Update{{Path: "name", Value: "Latest Fiber"}, {Path: "completedQuantity", Value: "60"}}); e != nil {
		t.Fatal(e)
	}
	page, e = svc.ListReferences(ctx, "member", gid, project.ID, "")
	if e != nil {
		t.Fatal(e)
	}
	for _, ref := range page.Items {
		if ref.SourceType == "object" && (ref.SourceTitle != "Latest Fiber" || ref.CompletedQuantity != "60") {
			t.Fatal("stale source", ref)
		}
	}
	after, e := before.Ref.Get(ctx)
	if e != nil || !reflect.DeepEqual(before.Data(), after.Data()) || !before.UpdateTime.Equal(after.UpdateTime) {
		t.Fatal("read wrote Project", e)
	}
	if _, e = svc.AddReference(ctx, "member", gid, project.ID, refs[0], 1); !errors.Is(e, p.ErrForbidden) {
		t.Fatal("duplicate owner bypass", e)
	}
	if _, e = svc.AddReference(ctx, "owner", gid, project.ID, p.Reference{SourceType: "subobject", SourceID: sid, ParentObjectID: strings.Repeat("1", 32)}, project.Version); e == nil {
		t.Fatal("cross parent")
	}
	if _, e = svc.AddReference(ctx, "owner", gid, project.ID, p.Reference{SourceType: "object", SourceID: strings.Repeat("1", 32)}, project.Version); e == nil {
		t.Fatal("foreign source")
	}

	// Old issue-only pagination advances on raw mixed rows, including short/empty pages.
	for n := range 21 {
		sourceID := fmt.Sprintf("%032x", n+1)
		if _, e = repository.Collection("objects").Doc(sourceID).Set(ctx, map[string]any{"schemaVersion": 1, "id": sourceID, "repositoryId": rid, "name": "Extra", "type": "non_work", "quantity": "1", "unit": "SET", "status": "not_started"}); e != nil {
			t.Fatal(e)
		}
		project, e = svc.AddReference(ctx, "owner", gid, project.ID, p.Reference{SourceType: "object", SourceID: sourceID}, project.Version)
		if e != nil {
			t.Fatal(e)
		}
	}
	cursor := ""
	issueSeen := 0
	for pages := 0; ; pages++ {
		if pages > 3 {
			t.Fatal("legacy cursor loop")
		}
		old, e := svc.ListItems(ctx, "owner", gid, project.ID, cursor)
		if e != nil {
			t.Fatal(e)
		}
		for _, i := range old.Items {
			if i.IssueID != iid {
				t.Fatal("nonissue in legacy list")
			}
			issueSeen++
		}
		if old.NextCursor == "" {
			break
		}
		if old.NextCursor == cursor {
			t.Fatal("nonadvancing cursor")
		}
		cursor = old.NextCursor
	}
	if issueSeen != 1 {
		t.Fatal("legacy issue lost/duplicated", issueSeen)
	}
	// Remove added rows again to keep subsequent exact counts focused on four source kinds.
	cursor = ""
	for {
		page, e := svc.ListReferences(ctx, "owner", gid, project.ID, cursor)
		if e != nil {
			t.Fatal(e)
		}
		for _, ref := range page.Items {
			if ref.SourceType == "object" && ref.SourceID != oid {
				project, e = svc.RemoveReference(ctx, "owner", gid, project.ID, ref.ID, project.Version)
				if e != nil {
					t.Fatal(e)
				}
			}
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if _, e = object.Update(ctx, []db.Update{{Path: "quantity", Value: "NaN"}}); e != nil {
		t.Fatal(e)
	}
	page, e = svc.ListReferences(ctx, "owner", gid, project.ID, "")
	if e != nil {
		t.Fatal(e)
	}
	for _, ref := range page.Items {
		if ref.SourceType == "object" && !ref.SourceUnavailable {
			t.Fatal("corrupt source progress exposed", ref)
		}
	}
	if _, e = object.Update(ctx, []db.Update{{Path: "quantity", Value: "140"}}); e != nil {
		t.Fatal(e)
	}
	if _, e = discussion.Delete(ctx); e != nil {
		t.Fatal(e)
	}
	page, e = svc.ListReferences(ctx, "owner", gid, project.ID, "")
	if e != nil {
		t.Fatal(e)
	}
	for _, ref := range page.Items {
		if ref.SourceType == "discussion" && (!ref.SourceUnavailable || ref.SourceTitle != "") {
			t.Fatal("missing source leak", ref)
		}
	}
	access.denied = true
	page, e = svc.ListReferences(ctx, "owner", gid, project.ID, "")
	if e != nil || len(page.Items) != 4 {
		t.Fatal(page, e)
	}
	for _, ref := range page.Items {
		if !ref.SourceUnavailable || ref.SourceTitle != "" || ref.SourceState != "" || ref.Quantity != "" {
			t.Fatal("revoked preview leak", ref)
		}
	}
	if _, e = svc.AddReference(ctx, "owner", gid, project.ID, refs[0], project.Version); !errors.Is(e, p.ErrNotFound) {
		t.Fatal("revoked duplicate", e)
	}
	version := project.Version
	project, e = svc.RemoveReference(ctx, "owner", gid, project.ID, ids["object"], version)
	if e != nil {
		t.Fatal("cannot remove revoked", e)
	}
	retry, e := svc.RemoveReference(ctx, "owner", gid, project.ID, ids["object"], version)
	if e != nil || retry.Version != project.Version {
		t.Fatal("remove retry", retry, e)
	}
	if _, e = object.Get(ctx); e != nil {
		t.Fatal("remove deleted source", e)
	}
	// HTTP read returns sealed unavailable rows and accepts no added snapshot/authority fields.
	mux := http.NewServeMux()
	projectHTTP.Register(mux, svc, auth{})
	path := fmt.Sprintf("/api/v1/spaces/%s/projects/%s/references", gid, project.ID)
	req := httptest.NewRequest("GET", path, nil)
	req.Header.Set("X-Test-Actor", "owner")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 200 || strings.Contains(w.Body.String(), "Latest Fiber") || strings.Contains(w.Body.String(), `"quantity"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	req = httptest.NewRequest("POST", path, strings.NewReader(fmt.Sprintf(`{"sourceType":"object","sourceId":"%s","expectedVersion":%d,"sourceTitle":"fake"}`, oid, project.Version)))
	req.Header.Set("X-Test-Actor", "owner")
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
}
