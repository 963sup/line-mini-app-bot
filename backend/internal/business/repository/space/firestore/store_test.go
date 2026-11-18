package firestore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/repository/space"
)

func TestEmulatorSpaceReceiptsScopeAndPagination(t *testing.T) {
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("requires explicit loopback emulator")
	}
	h, _, e := net.SplitHostPort(host)
	if e != nil || (h != "localhost" && !net.ParseIP(h).IsLoopback()) {
		t.Fatal("only loopback emulator allowed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	var random [8]byte
	if _, e = rand.Read(random[:]); e != nil {
		t.Fatal(e)
	}
	c, e := db.NewClient(ctx, "repository-space-test-"+hex.EncodeToString(random[:]))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
	s := &Store{Client: c, ChannelID: "isolated-space-test"}
	r := repository.Repository{ID: strings.Repeat("a", 32), OwnerID: "C" + strings.Repeat("b", 32), Status: "ACTIVE"}
	if _, e = s.repositoryRef(r.ID).Set(ctx, map[string]any{"id": r.ID, "ownerId": r.OwnerID, "status": "ACTIVE"}); e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make([]feature.Space, 3)
	errs := make([]error, 3)
	for i := range 3 {
		wg.Go(func() { results[i], errs[i] = s.Create(ctx, r, "actor", "same-space-key-12345", "一樓") })
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			t.Fatal(e)
		}
		if results[i].ID != results[0].ID {
			t.Fatal("duplicate spaces")
		}
	}
	if _, e = s.Create(ctx, r, "actor", "same-space-key-12345", "different"); !errors.Is(e, feature.ErrCreationConflict) {
		t.Fatalf("conflict=%v", e)
	}
	if _, e = s.Get(ctx, r, results[0].ID); e != nil {
		t.Fatal(e)
	}
	other := r
	other.ID = strings.Repeat("c", 32)
	if _, e = s.Get(ctx, other, results[0].ID); !errors.Is(e, feature.ErrNotFound) {
		t.Fatalf("cross-repo=%v", e)
	}
	fixtures := make([]feature.Space, 0, 20)
	for i := 0; i < 20; i++ {
		var id [16]byte
		if _, e := rand.Read(id[:]); e != nil {
			t.Fatal(e)
		}
		v := results[0]
		v.ID = hex.EncodeToString(id[:])
		fixtures = append(fixtures, v)
	}
	if e = c.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		for _, v := range fixtures {
			if err := tx.Create(s.repositoryRef(r.ID).Collection("spaces").Doc(v.ID), encode(v)); err != nil {
				return err
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	p, e := s.List(ctx, r, "actor", "")
	if e != nil || len(p.Items) != 20 || p.NextCursor == "" {
		t.Fatalf("page1=%+v err=%v", p, e)
	}
	second, e := s.List(ctx, r, "actor", p.NextCursor)
	if e != nil || len(second.Items) != 1 || second.NextCursor != "" {
		t.Fatalf("page2=%+v err=%v", second, e)
	}
	seen := map[string]bool{}
	for _, v := range append(p.Items, second.Items...) {
		if seen[v.ID] {
			t.Fatal("duplicate page item")
		}
		seen[v.ID] = true
	}
	if _, e = s.List(ctx, r, "other", p.NextCursor); !errors.Is(e, feature.ErrInput) {
		t.Fatalf("actor cursor=%v", e)
	}
	if _, e = s.List(ctx, other, "actor", p.NextCursor); !errors.Is(e, feature.ErrInput) {
		t.Fatalf("repo cursor=%v", e)
	}
	if _, e = s.repositoryRef(r.ID).Update(ctx, []db.Update{{Path: "status", Value: "INACTIVE"}}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Create(ctx, r, "actor", "same-space-key-12345", "一樓"); !errors.Is(e, feature.ErrNotFound) {
		t.Fatalf("inactive replay=%v", e)
	}
}
