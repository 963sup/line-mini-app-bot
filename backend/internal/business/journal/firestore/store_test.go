package firestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	groupdomain "github.com/963sup/line-mini-app-bot/internal/business/group"
	groupdb "github.com/963sup/line-mini-app-bot/internal/business/group/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/journal"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	repodb "github.com/963sup/line-mini-app-bot/internal/business/repository/firestore"
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
	if e := groups.Observe(ctx, groupdomain.Observation{GroupID: g, Status: "ACTIVE", At: time.Now(), EventID: "join", Users: []string{"writer"}}); e != nil {
		t.Fatal(e)
	}
	store := Store{Client: client, Channel: channel, Groups: &groupdomain.Service{Store: groups, Membership: membership{}}}
	draft := journal.Draft{GroupID: g, WorkDate: time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), Today: []journal.Work{{Name: "鋪設", Percent: 40}}}
	entry, e := store.Create(ctx, "writer", "key", draft)
	if e != nil {
		t.Fatal(e)
	}
	replay, e := store.Create(ctx, "writer", "key", draft)
	if e != nil || replay.ID != entry.ID || !entry.CreatedAt.Equal(replay.CreatedAt) {
		t.Fatalf("replay: %v", e)
	}
	draft.Today[0].Percent = 41
	if _, e := store.Create(ctx, "writer", "key", draft); !errors.Is(e, journal.ErrConflict) {
		t.Fatalf("expected conflict: %v", e)
	}
	if _, e := store.Get(ctx, "outsider", entry.ID); !errors.Is(e, journal.ErrNotFound) {
		t.Fatalf("unauthorized detail: %v", e)
	}
	if _, e := store.Create(ctx, "outsider", "other", draft); !errors.Is(e, journal.ErrNotFound) {
		t.Fatalf("unauthorized write: %v", e)
	}
	for i := 0; i < 20; i++ {
		if _, e := store.Create(ctx, "writer", fmt.Sprint(i), draft); e != nil {
			t.Fatal(e)
		}
	}
	f := journal.Filter{WorkDate: "2026-11-18"}
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
	if _, e := store.List(ctx, "writer", f); !errors.Is(e, journal.ErrInput) {
		t.Fatalf("filter changed: %v", e)
	}

	// A linked repository can become unavailable while the publishing group remains readable.
	rid := strings.Repeat("b", 32)
	repo := client.Collection("channels").Doc(channel).Collection("repositories").Doc(rid)
	_, e = repo.Set(ctx, map[string]any{"id": rid, "ownerId": g, "status": "ACTIVE", "version": int64(1)})
	if e != nil {
		t.Fatal(e)
	}
	store.Repositories = &repository.Service{Store: &repodb.Store{Client: client, ChannelID: channel}, Membership: membership{}}
	draft.Today[0].Reference = &journal.Reference{Kind: "repository", ID: rid}
	linked, e := store.Create(ctx, "writer", "linked", draft)
	if e != nil {
		t.Fatal(e)
	}
	_, e = repo.Update(ctx, []db.Update{{Path: "status", Value: "INACTIVE"}})
	if e != nil {
		t.Fatal(e)
	}
	redacted, e := store.Get(ctx, "writer", linked.ID)
	if e != nil || redacted.Today[0].Reference != nil {
		t.Fatalf("redaction: %v", e)
	}
	if _, e := store.Create(ctx, "writer", "denied-source", draft); !errors.Is(e, journal.ErrNotFound) {
		t.Fatalf("inactive source: %v", e)
	}
}
