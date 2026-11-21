package bootstrap

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
)

func TestNavigationLeaseSerializesAndProtectsNewOwner(t *testing.T) {
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if !strings.HasPrefix(host, "127.0.0.1:") && !strings.HasPrefix(host, "localhost:") {
		t.Skip("requires an isolated loopback Firestore emulator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := firestore.NewClient(ctx, "line-navigation-lease-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	channel := time.Now().Format("150405.000000000")
	release, err := acquireNavigationLease(ctx, db, channel)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = release() })
	contender, cancelContender := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancelContender()
	if _, err := acquireNavigationLease(contender, db, channel); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("overlapping holder accepted: %v", err)
	}
	ref := db.Collection("line_navigation_publication_leases").Doc(channel)
	if _, err := ref.Update(ctx, []firestore.Update{{Path: "expiresAt", Value: time.Now().Add(-time.Minute)}}); err != nil {
		t.Fatal(err)
	}
	releaseNew, err := acquireNavigationLease(ctx, db, channel)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = releaseNew() }()
	if err := release(); err != nil {
		t.Fatal(err)
	}
	if _, err := ref.Get(ctx); err != nil {
		t.Fatalf("old owner deleted new holder: %v", err)
	}
}
