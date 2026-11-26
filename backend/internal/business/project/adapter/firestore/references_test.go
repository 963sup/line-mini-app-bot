package firestore_test

import (
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

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

	orgdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	projectDB "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	projectHTTP "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/http"
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
	gid, rid := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", strings.Repeat("b", 32)
	if _, e = root.Collection("organizations").Doc(gid).Set(ctx, map[string]any{"ownerUserId": "owner", "memberUserIds": []string{"member"}}); e != nil {
		t.Fatal(e)
	}
	repository := root.Collection("repositories").Doc(rid)
	if _, e = repository.Set(ctx, map[string]any{"id": rid, "ownerId": gid, "status": "ACTIVE", "version": int64(1)}); e != nil {
		t.Fatal(e)
	}
	oid, sid, iid, did := strings.Repeat("c", 32), strings.Repeat("d", 32), strings.Repeat("e", 32), strings.Repeat("f", 32)
	object := repository.Collection("objects").Doc(oid)
	sub := repository.Collection("objects").Doc(sid)
	issue := repository.Collection("issues").Doc(iid)
	discussion := repository.Collection("discussions").Doc(did)
	data := []struct {
		ref    *db.DocumentRef
		fields map[string]any
	}{{object, map[string]any{"schemaVersion": 1, "id": oid, "repositoryId": rid, "name": "Fiber", "type": "work", "quantity": "140", "unit": "M", "completedQuantity": "40", "status": "in_progress"}}, {sub, map[string]any{"schemaVersion": 1, "id": sid, "repositoryId": rid, "parentId": oid, "name": "Splice", "type": "work", "quantity": "11", "unit": "SET", "completedQuantity": "0", "status": "not_started"}}, {issue, map[string]any{"schemaVersion": 1, "id": iid, "repositoryId": rid, "title": "Question", "state": "completed"}}, {discussion, map[string]any{"schemaVersion": 1, "id": did, "repositoryId": rid, "title": "Discussion"}}}
	for _, source := range data {
		if _, e = source.ref.Set(ctx, source.fields); e != nil {
			t.Fatal(e)
		}
	}
	access := &sourceAccess{root: root}
	store := &projectDB.Store{Client: client, ChannelID: root.ID, OrganizationVisible: (orgdb.Store{}).VisibleTo}
	svc := &projectApplication.Service{AuthorizeOrganization: store.AuthorizeOrganization, Sources: access, Store: store}
	project, e := svc.Create(ctx, "owner", gid, "typed-reference-project", projectDomainEntity.Content{Title: "Existing project", SourceRepositoryID: rid})
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
	refs := []projectDomainEntity.Reference{{SourceType: "issue", SourceID: iid}, {SourceType: "object", SourceID: oid}, {SourceType: "object", SourceID: sid}, {SourceType: "discussion", SourceID: did}}
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
		key := ref.SourceType
		if ref.SourceID == sid {
			key = "child"
		}
		ids[key] = ref.ID
		if ref.SourceUnavailable || ref.SourceTitle == "" {
			t.Fatal(ref)
		}
	}
	if ids["issue"] != oldIssueID {
		t.Fatal("duplicate legacy issue")
	}
	raw, e := root.Collection("organizations").Doc(gid).Collection("projects").Doc(project.ID).Collection("items").Documents(ctx).GetAll()
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
	before, e := root.Collection("organizations").Doc(gid).Collection("projects").Doc(project.ID).Get(ctx)
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
		if ref.SourceID == oid && (ref.SourceTitle != "Latest Fiber" || ref.CompletedQuantity != "60") {
			t.Fatal("stale source", ref)
		}
	}
	after, e := before.Ref.Get(ctx)
	if e != nil || !reflect.DeepEqual(before.Data(), after.Data()) || !before.UpdateTime.Equal(after.UpdateTime) {
		t.Fatal("read wrote Project", e)
	}
	if _, e = svc.AddReference(ctx, "member", gid, project.ID, refs[0], 1); !errors.Is(e, projectDomainEntity.ErrForbidden) {
		t.Fatal("duplicate owner bypass", e)
	}
	if _, e = svc.AddReference(ctx, "owner", gid, project.ID, projectDomainEntity.Reference{SourceType: "subobject", SourceID: sid}, project.Version); e == nil {
		t.Fatal("obsolete source type was accepted")
	}
	if _, e = svc.AddReference(ctx, "owner", gid, project.ID, projectDomainEntity.Reference{SourceType: "object", SourceID: strings.Repeat("1", 32)}, project.Version); e == nil {
		t.Fatal("foreign source")
	}

	// Old issue-only pagination advances on raw mixed rows, including short/empty pages.
	for n := range 21 {
		sourceID := fmt.Sprintf("%032x", n+1)
		if _, e = repository.Collection("objects").Doc(sourceID).Set(ctx, map[string]any{"schemaVersion": 1, "id": sourceID, "repositoryId": rid, "name": "Extra", "type": "non_work", "quantity": "1", "unit": "SET", "status": "not_started"}); e != nil {
			t.Fatal(e)
		}
		project, e = svc.AddReference(ctx, "owner", gid, project.ID, projectDomainEntity.Reference{SourceType: "object", SourceID: sourceID}, project.Version)
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
	// Remove pagination fixtures while keeping the four original references.
	cursor = ""
	for {
		page, e := svc.ListReferences(ctx, "owner", gid, project.ID, cursor)
		if e != nil {
			t.Fatal(e)
		}
		for _, ref := range page.Items {
			if ref.SourceType == "object" && ref.SourceID != oid && ref.SourceID != sid {
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
		if ref.SourceType == "object" && ref.SourceID == oid && !ref.SourceUnavailable {
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
	// A second group may use the same issue ID. Its link and read authority are independent.
	secondID := strings.Repeat("9", 32)
	second := root.Collection("repositories").Doc(secondID)
	if _, e = second.Set(ctx, map[string]any{"id": secondID, "ownerId": "C" + secondID, "status": "ACTIVE", "version": 1}); e != nil {
		t.Fatal(e)
	}
	if _, e = second.Collection("issues").Doc(iid).Set(ctx, map[string]any{"schemaVersion": 1, "id": iid, "repositoryId": secondID, "title": "Second group issue", "state": "open"}); e != nil {
		t.Fatal(e)
	}
	secondRef := projectDomainEntity.Reference{SourceRepositoryID: secondID, SourceType: "issue", SourceID: iid}
	if _, e = svc.AddReference(ctx, "owner", gid, project.ID, secondRef, project.Version); !errors.Is(e, projectDomainEntity.ErrForbidden) {
		t.Fatalf("unlinked source: %v", e)
	}
	project, e = svc.Update(ctx, "owner", gid, project.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: project.Title, Body: project.Body}, ExpectedVersion: project.Version, LinkRepositoryID: secondID})
	if e != nil {
		t.Fatal(e)
	}
	project, e = svc.AddReference(ctx, "owner", gid, project.ID, secondRef, project.Version)
	if e != nil {
		t.Fatal(e)
	}
	page, e = svc.ListReferences(ctx, "owner", gid, project.ID, "")
	if e != nil {
		t.Fatal(e)
	}
	secondItem := ""
	for _, ref := range page.Items {
		if ref.SourceRepositoryID == secondID && ref.SourceTitle == "Second group issue" {
			secondItem = ref.ID
		}
	}
	if secondItem == "" || secondItem == oldIssueID {
		t.Fatal("distinct second issue missing", page)
	}
	other, e := svc.Create(ctx, "owner", gid, "independent-issue-board", projectDomainEntity.Content{Title: "Audit", SourceRepositoryID: rid})
	if e != nil {
		t.Fatal(e)
	}
	other, e = svc.Update(ctx, "owner", gid, other.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: other.Title}, ExpectedVersion: other.Version, LinkRepositoryID: secondID})
	if e != nil {
		t.Fatal(e)
	}
	other, e = svc.AddReference(ctx, "owner", gid, other.ID, secondRef, other.Version)
	if e != nil {
		t.Fatal(e)
	}
	project, e = svc.Update(ctx, "owner", gid, project.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: project.Title, Body: project.Body}, ExpectedVersion: project.Version, ItemID: secondItem, ItemStatus: "done"})
	if e != nil {
		t.Fatal(e)
	}
	for board, want := range map[string]string{project.ID: "done", other.ID: "todo"} {
		items, e := svc.ListReferences(ctx, "owner", gid, board, "")
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, item := range items.Items {
			if item.SourceRepositoryID == secondID && item.SourceID == iid {
				found = true
				if item.ItemStatus != want || item.SourceState != "open" {
					t.Fatalf("board %s: %+v", board, item)
				}
			}
		}
		if !found {
			t.Fatal("missing issue on board", board)
		}
	}
	project, e = svc.Update(ctx, "owner", gid, project.ID, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: project.Title, Body: project.Body}, ExpectedVersion: project.Version, UnlinkRepositoryID: secondID})
	if e != nil {
		t.Fatal(e)
	}
	page, e = svc.ListReferences(ctx, "owner", gid, project.ID, "")
	if e != nil {
		t.Fatal(e)
	}
	for _, ref := range page.Items {
		if ref.ID == secondItem && (!ref.SourceUnavailable || ref.SourceTitle != "") {
			t.Fatal("unlinked preview", ref)
		}
	}
	project, e = svc.RemoveReference(ctx, "owner", gid, project.ID, secondItem, project.Version)
	if e != nil {
		t.Fatal(e)
	}
	// Submission evidence belongs to the existing reference; submitting never edits the source.
	if _, e = sub.Update(ctx, []db.Update{{Path: "version", Value: int64(7)}}); e != nil {
		t.Fatal(e)
	}
	beforeSource, e := sub.Get(ctx)
	if e != nil {
		t.Fatal(e)
	}
	note := "配線檢查完成"
	sent, e := svc.Inspections(ctx, "member", gid, project.ID, ids["child"], &note)
	if e != nil || len(sent) != 1 || sent[0].SourceVersion != 7 || sent[0].Status != "pending" {
		t.Fatal(sent, e)
	}
	retried, e := svc.Inspections(ctx, "member", gid, project.ID, ids["child"], &note)
	if e != nil || len(retried) != 1 || retried[0].ID != sent[0].ID {
		t.Fatal(retried, e)
	}
	if _, e = svc.Inspections(ctx, "owner", gid, project.ID, ids["child"], &note); !errors.Is(e, projectDomainEntity.ErrInspectionPending) {
		t.Fatal(e)
	}
	if _, e = svc.RemoveReference(ctx, "owner", gid, project.ID, ids["child"], project.Version); !errors.Is(e, projectDomainEntity.ErrInspectionHistory) {
		t.Fatal(e)
	}
	afterSource, e := sub.Get(ctx)
	if e != nil || !beforeSource.UpdateTime.Equal(afterSource.UpdateTime) {
		t.Fatal("submission modified source", e)
	}
	if _, e = sub.Delete(ctx); e != nil {
		t.Fatal(e)
	}
	history, e := svc.Inspections(ctx, "member", gid, project.ID, ids["child"], nil)
	if e != nil || len(history) != 1 || history[0].ID != sent[0].ID {
		t.Fatal("deleted source hid authorized history", history, e)
	}
	if _, e = svc.Inspections(ctx, "member", gid, project.ID, ids["child"], &note); e == nil {
		t.Fatal("submitted on deleted source")
	}
	access.denied = true
	if _, e = svc.Inspections(ctx, "member", gid, project.ID, ids["child"], nil); !errors.Is(e, projectDomainEntity.ErrNotFound) {
		t.Fatal("revoked inspection readable", e)
	}
	page, e = svc.ListReferences(ctx, "owner", gid, project.ID, "")
	if e != nil || len(page.Items) != 4 {
		t.Fatal(page, e)
	}
	for _, ref := range page.Items {
		if !ref.SourceUnavailable || ref.SourceTitle != "" || ref.SourceState != "" || ref.Quantity != "" {
			t.Fatal("revoked preview leak", ref)
		}
	}
	if _, e = svc.AddReference(ctx, "owner", gid, project.ID, refs[0], project.Version); !errors.Is(e, projectDomainEntity.ErrNotFound) {
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
	path := fmt.Sprintf("/api/v1/owners/%s/projects/%s/references", gid, project.ID)
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
