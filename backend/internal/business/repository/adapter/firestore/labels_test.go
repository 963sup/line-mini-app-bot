package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
)

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"testing"

	db "cloud.google.com/go/firestore"
)

func TestRenameLabelPreservesIdentityAndCreationReplay(t *testing.T) {
	addr := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if addr == "" {
		t.Skip("requires loopback Firestore emulator")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("requires loopback")
	}
	ctx := context.Background()
	c, err := db.NewClient(ctx, "labels-local")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
	s := Store{Client: c, ChannelID: c.Collection("channels").NewDoc().ID}
	r := repositoryDomainEntity.Repository{ID: strings.Repeat("1", 32), OwnerType: repositoryDomainEntity.OwnerUser, OwnerID: "U" + strings.Repeat("2", 32), Status: "ACTIVE", Version: 1}
	if _, err = s.repositoryRef(r.ID).Set(ctx, document{ID: r.ID, OwnerType: string(r.OwnerType), OwnerID: r.OwnerID, Status: r.Status, Version: 1}); err != nil {
		t.Fatal(err)
	}
	create := func(key, name string) repositoryDomainEntityLabels.Label {
		l, e := s.CreateLabel(ctx, r, r.OwnerID, key, name)
		if e != nil {
			t.Fatal(e)
		}
		return l
	}
	first := create("first-create", "old")
	create("second-create", "occupied")
	if _, err = s.RenameLabel(ctx, r, r.OwnerID, first.ID, "old", "Occupied"); !errors.Is(err, repositoryDomainEntityLabels.ErrNameConflict) {
		t.Fatal(err)
	}
	if _, err = s.RenameLabel(ctx, r, r.OwnerID, first.ID, "old", "new"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RenameLabel(ctx, r, r.OwnerID, first.ID, "old", "other"); !errors.Is(err, repositoryDomainEntityLabels.ErrStaleName) {
		t.Fatal(err)
	}
	create("third-create", "old")
	replay := create("first-create", "old")
	if replay.ID != first.ID || replay.Name != "new" {
		t.Fatal(replay)
	}
}
