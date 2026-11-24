package firestore_test

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"

	spaceDB "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore/group"
	spaceLine "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/line"
	groupport "github.com/963sup/line-mini-app-bot/internal/business/team/application/port"
	groupevent "github.com/963sup/line-mini-app-bot/internal/business/team/domain/event"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGroupIdentityAndLifecycleWithoutRepository(t *testing.T) {
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
	channel := "space-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	root := client.Collection("channels").Doc(channel)
	store := &spaceDB.Store{Client: client, ChannelID: channel}
	at := time.Now().UTC().Truncate(time.Millisecond)
	gid := "C" + strings.Repeat("a", 32)
	event := groupevent.Observation{GroupID: gid, Name: "Independent Group", Status: "ACTIVE", EventID: "join", At: at, Users: []string{"actor"}}
	if err = store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	first, err := store.GetByID(ctx, gid)
	if err != nil || first.ID != gid || first.GroupID != gid || first.Status != "ACTIVE" {
		t.Fatalf("Group derives channel/group identity: %+v %v", first, err)
	}
	for _, ref := range []*db.DocumentRef{root.Collection("repositories").Doc(gid), root.Collection("repositoryOwnerBindings").Doc(gid)} {
		if _, err = ref.Get(ctx); status.Code(err) != codes.NotFound {
			t.Fatalf("Group must not create or depend on Repository: %s %v", ref.Path, err)
		}
	}
	observer := &spaceLine.Observer{Store: store, GroupName: func(context.Context, string) string { return "Independent Group" }}
	group := transport.GroupEvent{GroupID: gid, EventID: "next-message", Kind: transport.GroupMessage, At: at.Add(10 * time.Millisecond), Users: []string{"actor"}}
	if err = observer.Observe(ctx, nil, &group); err != nil {
		t.Fatalf("nonjoin event must update independent Group: %v", err)
	}
	if err = store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	again, err := store.GetByID(ctx, gid)
	if err != nil || again.Version != first.Version {
		t.Fatalf("replay changed independent identity/version: %+v %v", again, err)
	}
	otherChannel := &spaceDB.Store{Client: client, ChannelID: channel + "-other"}
	otherEvent := event
	otherEvent.Name = "Other channel Group"
	if err = otherChannel.Observe(ctx, otherEvent); err != nil {
		t.Fatal(err)
	}
	sameGroupOtherChannel, err := otherChannel.GetByID(ctx, gid)
	if err != nil || sameGroupOtherChannel.Name != "Other channel Group" {
		t.Fatalf("group identity must remain channel-scoped: %+v %v", sameGroupOtherChannel, err)
	}
	unchanged, err := store.GetByID(ctx, gid)
	if err != nil || unchanged.Name != "Independent Group" {
		t.Fatalf("another channel changed Group state: %+v %v", unchanged, err)
	}
	page, err := store.Candidates(ctx, "actor", "")
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != gid {
		t.Fatalf("independent candidate discovery: %+v %v", page, err)
	}
	event.Status = "INACTIVE"
	event.At = at.Add(time.Second)
	event.EventID = "leave"
	if err = store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	event.Status = "ACTIVE"
	event.EventID = "stale-active"
	event.At = at
	if err = store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	inactive, err := store.GetByID(ctx, gid)
	if err != nil || inactive.Status != "INACTIVE" {
		t.Fatalf("stale group event reactivated Group: %+v %v", inactive, err)
	}
	other := "C" + strings.Repeat("c", 32)
	orphan := root.Collection("groups").Doc(other)
	if _, err = orphan.Set(ctx, map[string]any{"id": other, "groupId": gid, "name": "do not overwrite", "status": "ACTIVE", "version": int64(7), "createdAt": at, "updatedAt": at}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.GetByID(ctx, other); !errors.Is(err, groupport.ErrUnavailable) {
		t.Fatalf("stored group/Group mismatch must not authorize: %v", err)
	}
	event.GroupID = other
	event.EventID = "orphan-collision"
	event.At = at.Add(2 * time.Second)
	if err = store.Observe(ctx, event); !errors.Is(err, groupport.ErrUnavailable) {
		t.Fatalf("orphan identity collision must not overwrite: %v", err)
	}
	snapshot, err := orphan.Get(ctx)
	if err != nil || snapshot.Data()["version"] != int64(7) {
		t.Fatalf("collision mutated existing Group: %v %v", snapshot, err)
	}

	incompleteID := "C" + strings.Repeat("d", 32)
	incomplete := root.Collection("groups").Doc(incompleteID)
	if _, err = incomplete.Set(ctx, map[string]any{"id": incompleteID, "status": "ACTIVE", "version": int64(7)}); err != nil {
		t.Fatal(err)
	}
	event.GroupID = incompleteID
	event.EventID = "incomplete-identity"
	if _, err = store.GetByID(ctx, incompleteID); !errors.Is(err, groupport.ErrUnavailable) {
		t.Fatalf("missing stored groupId must not authorize: %v", err)
	}
	if err = store.Observe(ctx, event); !errors.Is(err, groupport.ErrUnavailable) {
		t.Fatalf("observation must reject a missing stored groupId: %v", err)
	}
	snapshot, err = incomplete.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := snapshot.Data()["groupId"]; present {
		t.Fatal("observation populated the missing stored groupId")
	}
	if _, err = root.Collection("teams").Doc(incompleteID).Get(ctx); status.Code(err) != codes.NotFound {
		t.Fatalf("invalid observation must not create a Team: %v", err)
	}
}
