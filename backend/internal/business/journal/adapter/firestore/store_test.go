package firestore

import (
	journalquery "github.com/963sup/line-mini-app-bot/internal/business/journal/application/query"
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	projectgroup "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/group"
	projectsources "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/repository"
	projectapp "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
)

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"

	groupdb "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/firestore"
	groupquery "github.com/963sup/line-mini-app-bot/internal/linegroup/application/query"
	groupevent "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/event"

	journaldto "github.com/963sup/line-mini-app-bot/internal/business/journal/application/dto"
	journalentity "github.com/963sup/line-mini-app-bot/internal/business/journal/domain/entity"

	repodb "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore"
)

type membership struct{}

func (membership) IsMember(_ context.Context, _, actor string) (bool, error) {
	return actor == "writer", nil
}

func TestPublishingAndScopedPagination(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires Firestore emulator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	client, e := db.NewClient(ctx, "journal-local")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Firestore client: %v", err)
		}
	})
	channel := fmt.Sprintf("test-%d", time.Now().UnixNano())
	groups := &groupdb.Store{Client: client, ChannelID: channel}
	g := "C" + strings.Repeat("a", 32)
	if e := groups.Observe(ctx, groupevent.Observation{GroupID: g, Status: "ACTIVE", At: time.Now(), EventID: "join", Users: []string{"writer"}}); e != nil {
		t.Fatal(e)
	}
	store := Store{Client: client, Channel: channel, ReferenceCatalog: journalquery.ReferenceCatalog{Groups: &groupquery.Service{Store: groups, Membership: membership{}}}}
	rid := strings.Repeat("b", 32)
	repo := client.Collection("channels").Doc(channel).Collection("repositories").Doc(rid)
	_, e = repo.Set(ctx, map[string]any{"id": rid, "ownerId": g, "status": "ACTIVE", "version": int64(1)})
	if e != nil {
		t.Fatal(e)
	}
	store.Repositories = &repositoryApplication.Service{Store: &repodb.Store{Client: client, ChannelID: channel}, Membership: membership{}}
	draft := journalentity.Draft{RepositoryID: rid, WorkDate: time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), Today: []journalentity.Work{{Name: "鋪設", Percent: 40}}}
	entry, e := store.Create(ctx, "writer", "key", draft)
	if e != nil {
		t.Fatal(e)
	}
	replay, e := store.Create(ctx, "writer", "key", draft)
	if e != nil || replay.ID != entry.ID || !entry.CreatedAt.Equal(replay.CreatedAt) {
		t.Fatalf("replay: %v", e)
	}
	draft.Today[0].Percent = 41
	if _, e := store.Create(ctx, "writer", "key", draft); !errors.Is(e, journalentity.ErrConflict) {
		t.Fatalf("expected conflict: %v", e)
	}
	if _, e := store.Get(ctx, "outsider", entry.ID); !errors.Is(e, journalentity.ErrNotFound) {
		t.Fatalf("unauthorized detail: %v", e)
	}
	if _, e := store.Create(ctx, "outsider", "other", draft); !errors.Is(e, journalentity.ErrNotFound) {
		t.Fatalf("unauthorized write: %v", e)
	}
	for i := 0; i < 20; i++ {
		if _, e := store.Create(ctx, "writer", fmt.Sprint(i), draft); e != nil {
			t.Fatal(e)
		}
	}
	f := journaldto.Filter{RepositoryID: rid, WorkDate: "2026-11-18"}
	p, e := store.List(ctx, "writer", f)
	if e != nil || len(p.Items) != 20 || p.NextCursor == "" {
		t.Fatalf("page: %d %v", len(p.Items), e)
	}
	f.Cursor = p.NextCursor
	p, e = store.List(ctx, "writer", f)
	if e != nil || len(p.Items) != 1 || p.Items[0].ID != entry.ID {
		t.Fatalf("continuation: %d %v", len(p.Items), e)
	}
	f.GroupID = g
	if _, e := store.List(ctx, "writer", f); !errors.Is(e, journalentity.ErrInput) {
		t.Fatalf("filter changed: %v", e)
	}

	// A linked repository can become unavailable while the publishing group remains readable.
	store.Projects = &projectapp.Service{Groups: projectgroup.Access{Groups: &groupquery.Service{Store: groups, Membership: membership{}}}, Store: &projectdb.Store{Client: client, ChannelID: channel}, Sources: projectsources.Sources{Repositories: store.Repositories}}
	draft.Today[0].Reference = &journalentity.Reference{Kind: "repository", ID: rid}
	objectID, childID, spaceID := strings.Repeat("c", 32), strings.Repeat("d", 32), strings.Repeat("e", 32)
	object := repo.Collection("objects").Doc(objectID)
	child := repo.Collection("objects").Doc(childID)
	space := repo.Collection("spaces").Doc(spaceID)
	for ref, data := range map[*db.DocumentRef]map[string]any{
		object: {"id": objectID, "repositoryId": rid, "name": "父物件"},
		child:  {"id": childID, "repositoryId": rid, "parentId": objectID, "name": "子物件"},
		space:  {"id": spaceID, "repositoryId": rid, "name": "作業空間"},
	} {
		if _, e := ref.Set(ctx, data); e != nil {
			t.Fatal(e)
		}
	}
	draft.Today[0].ObjectID, draft.Today[0].SpaceIDs = childID, []string{spaceID}
	draft.Tomorrow = []journalentity.Work{{Name: "明日工作", Percent: 60, Reference: &journalentity.Reference{Kind: "repository", ID: rid}, ObjectID: objectID, SpaceIDs: []string{spaceID}}}
	linked, e := store.Create(ctx, "writer", "linked", draft)
	if e != nil {
		t.Fatal(e)
	}
	assertSnapshot := func(entry journalentity.Entry) {
		t.Helper()
		if len(entry.Today) != 1 || len(entry.Tomorrow) != 1 {
			t.Fatalf("missing work rows: %+v", entry)
		}
		row := entry.Today[0]
		if row.ObjectID != childID || row.ObjectName != "子物件" || len(row.SpaceIDs) != 1 || row.SpaceIDs[0] != spaceID || len(row.SpaceNames) != 1 || row.SpaceNames[0] != "作業空間" {
			t.Fatalf("wrong today snapshot: %+v", row)
		}
		if row := entry.Tomorrow[0]; row.ObjectName != "父物件" || len(row.SpaceNames) != 1 || row.SpaceNames[0] != "作業空間" {
			t.Fatalf("wrong tomorrow snapshot: %+v", row)
		}
	}
	assertSnapshot(linked)
	if draft.Today[0].ObjectName != "" || len(draft.Today[0].SpaceNames) != 0 || draft.Tomorrow[0].ObjectName != "" {
		t.Fatal("publishing mutated the caller's retry payload")
	}
	for _, ref := range []*db.DocumentRef{object, child, space} {
		if _, e := ref.Update(ctx, []db.Update{{Path: "name", Value: "已改名"}}); e != nil {
			t.Fatal(e)
		}
	}
	replayed, e := store.Create(ctx, "writer", "linked", draft)
	if e != nil || replayed.ID != linked.ID || !replayed.CreatedAt.Equal(linked.CreatedAt) {
		t.Fatalf("snapshot replay: %v", e)
	}
	assertSnapshot(replayed)
	persisted, e := store.Get(ctx, "writer", linked.ID)
	if e != nil {
		t.Fatal(e)
	}
	assertSnapshot(persisted)
	otherRepo := repo.Parent.Doc(strings.Repeat("2", 32))
	if _, e := otherRepo.Set(ctx, map[string]any{"id": otherRepo.ID, "ownerId": "C" + strings.Repeat("3", 32), "status": "ACTIVE", "version": int64(1)}); e != nil {
		t.Fatal(e)
	}
	entriesBefore, e := store.repositoryEntries(rid).Documents(ctx).GetAll()
	if e != nil {
		t.Fatal(e)
	}
	receiptsBefore, e := store.root().Collection("journalReceipts").Documents(ctx).GetAll()
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		name string
		row  journalentity.Work
		want error
	}{
		{"wrong-group", journalentity.Work{Reference: &journalentity.Reference{Kind: "repository", ID: otherRepo.ID}, ObjectID: objectID}, journalentity.ErrNotFound},
		{"missing-object", journalentity.Work{Reference: &journalentity.Reference{Kind: "repository", ID: rid}, ObjectID: strings.Repeat("f", 32)}, journalentity.ErrNotFound},
		{"duplicate-spaces", journalentity.Work{Reference: &journalentity.Reference{Kind: "repository", ID: rid}, SpaceIDs: []string{spaceID, spaceID}}, journalentity.ErrInput},
		{"missing-space", journalentity.Work{Reference: &journalentity.Reference{Kind: "repository", ID: rid}, ObjectID: objectID, SpaceIDs: []string{strings.Repeat("4", 32)}}, journalentity.ErrNotFound},
	} {
		invalid := draft
		invalid.Today = []journalentity.Work{tc.row}
		if _, e := store.Create(ctx, "writer", tc.name, invalid); !errors.Is(e, tc.want) {
			t.Fatalf("%s: got %v, want %v", tc.name, e, tc.want)
		}
	}
	entriesAfter, e := store.repositoryEntries(rid).Documents(ctx).GetAll()
	if e != nil {
		t.Fatal(e)
	}
	receiptsAfter, e := store.root().Collection("journalReceipts").Documents(ctx).GetAll()
	if e != nil || len(entriesAfter) != len(entriesBefore) || len(receiptsAfter) != len(receiptsBefore) {
		t.Fatalf("rejected source left an entry or receipt: entries %d -> %d, receipts %d -> %d, err %v", len(entriesBefore), len(entriesAfter), len(receiptsBefore), len(receiptsAfter), e)
	}
	legacy := linked
	legacy.ID = strings.Repeat("9", 32)
	legacy.RepositoryID, legacy.GroupID = "", g
	payload, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.entries(g).Doc(legacy.ID).Set(ctx, record{Payload: string(payload), Order: "2026-11-18/legacy"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.root().Collection("journalReceipts").Doc(legacy.ID).Set(ctx, receipt{Group: g}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(ctx, "writer", "legacy-rejected", legacy.Draft); !errors.Is(err, journalentity.ErrInput) {
		t.Fatalf("legacy write accepted: %v", err)
	}
	page, err := store.List(ctx, "writer", journaldto.Filter{GroupID: g, WorkDate: "2026-11-18"})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != legacy.ID {
		t.Fatalf("legacy list failed: %+v %v", page, err)
	}
	_, e = repo.Update(ctx, []db.Update{{Path: "status", Value: "INACTIVE"}})
	if e != nil {
		t.Fatal(e)
	}
	if _, err := store.Get(ctx, "writer", linked.ID); !errors.Is(err, journalentity.ErrNotFound) {
		t.Fatalf("inactive repository still readable: %v", err)
	}
	redacted, e := store.Get(ctx, "writer", legacy.ID)
	if e != nil || redacted.Today[0].Reference != nil {
		t.Fatalf("redaction: %v", e)
	}
	for _, row := range append(redacted.Today, redacted.Tomorrow...) {
		if row.Reference != nil || row.ObjectID != "" || len(row.SpaceIDs) != 0 || row.ObjectName != "" || len(row.SpaceNames) != 0 {
			t.Fatalf("unreadable source leaked selection or names: %+v", row)
		}
	}
	if _, e := store.Create(ctx, "writer", "denied-source", draft); !errors.Is(e, journalentity.ErrNotFound) {
		t.Fatalf("inactive source: %v", e)
	}
}
