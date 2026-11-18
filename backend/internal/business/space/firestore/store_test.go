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
	sp "github.com/963sup/line-mini-app-bot/internal/business/space"
	spaceDB "github.com/963sup/line-mini-app-bot/internal/business/space/firestore"
	spaceLine "github.com/963sup/line-mini-app-bot/internal/business/space/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSpaceIdentityAndLifecycleWithoutRepository(t *testing.T) {
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
	event := sp.Observation{GroupID: gid, Name: "Independent Space", Status: "ACTIVE", EventID: "join", At: at, Users: []string{"actor"}}
	if err = store.Observe(ctx, event); err != nil {
		t.Fatal(err)
	}
	first, err := store.GetByID(ctx, gid)
	if err != nil || first.ID != gid || first.GroupID != gid || first.Status != "ACTIVE" {
		t.Fatalf("Space derives channel/group identity: %+v %v", first, err)
	}
	for _, ref := range []*db.DocumentRef{root.Collection("repositories").Doc(gid), root.Collection("repositoryGroupBindings").Doc(gid)} {
		if _, err = ref.Get(ctx); status.Code(err) != codes.NotFound {
			t.Fatalf("Space must not create or depend on Repository: %s %v", ref.Path, err)
		}
	}
	byGroup, err := store.GetByGroupID(ctx, gid)
	if err != nil || byGroup.ID != gid {
		t.Fatalf("Space lookup cannot use Repository group binding: %+v %v", byGroup, err)
	}
	observer := &spaceLine.Observer{Store: store, GroupName: func(context.Context, string) string { return "Independent Space" }}
	if _, handled, e := observer.Observe(ctx, webhook.MessageEvent{Source: webhook.GroupSource{GroupId: gid, UserId: "actor"}, Timestamp: at.Add(10 * time.Millisecond).UnixMilli(), WebhookEventId: "next-message", Mode: "active"}); e != nil || handled {
		t.Fatalf("nonjoin event must update independent Space and remain nonconsuming: handled=%v err=%v", handled, e)
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
	otherEvent.Name = "Other channel Space"
	if err = otherChannel.Observe(ctx, otherEvent); err != nil {
		t.Fatal(err)
	}
	sameGroupOtherChannel, err := otherChannel.GetByID(ctx, gid)
	if err != nil || sameGroupOtherChannel.Name != "Other channel Space" {
		t.Fatalf("group identity must remain channel-scoped: %+v %v", sameGroupOtherChannel, err)
	}
	unchanged, err := store.GetByID(ctx, gid)
	if err != nil || unchanged.Name != "Independent Space" {
		t.Fatalf("another channel changed Space state: %+v %v", unchanged, err)
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
		t.Fatalf("stale group event reactivated Space: %+v %v", inactive, err)
	}
	// A deterministic group identity must reject conflicting bindings and orphan roots.
	if _, err = root.Collection("spaceGroupBindings").Doc(gid).Set(ctx, map[string]any{"spaceId": "C" + strings.Repeat("b", 32)}); err != nil {
		t.Fatal(err)
	}
	if err = store.Observe(ctx, event); !errors.Is(err, sp.ErrUnavailable) {
		t.Fatalf("conflicting group binding must fail closed: %v", err)
	}
	other := "C" + strings.Repeat("c", 32)
	orphan := root.Collection("spaces").Doc(other)
	if _, err = orphan.Set(ctx, map[string]any{"id": other, "groupId": gid, "name": "do not overwrite", "status": "ACTIVE", "version": int64(7), "createdAt": at, "updatedAt": at}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.GetByID(ctx, other); !errors.Is(err, sp.ErrUnavailable) {
		t.Fatalf("stored group/Space mismatch must not authorize: %v", err)
	}
	event.GroupID = other
	event.EventID = "orphan-collision"
	event.At = at.Add(2 * time.Second)
	if err = store.Observe(ctx, event); !errors.Is(err, sp.ErrUnavailable) {
		t.Fatalf("orphan identity collision must not overwrite: %v", err)
	}
	snapshot, err := orphan.Get(ctx)
	if err != nil || snapshot.Data()["version"] != int64(7) {
		t.Fatalf("collision mutated existing Space: %v %v", snapshot, err)
	}
}
