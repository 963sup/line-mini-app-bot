package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	db "cloud.google.com/go/firestore"
)

func cliArgs() []string {
	return []string{"-project", "migration-test", "-database", "(default)", "-channel", "123", "-repository", strings.Repeat("a", 32), "-manifest", "/tmp/manifest.json"}
}

func TestExplicitScopeAndWriterReadiness(t *testing.T) {
	if _, err := parse(nil); err == nil {
		t.Fatal("missing explicit scope accepted")
	}
	o, err := parse(cliArgs())
	if err != nil || o.Apply || o.Rollback || o.ReleaseLock {
		t.Fatal("default must be read-only dry-run", err)
	}
	for _, mode := range []string{"-apply", "-rollback", "-release-lock"} {
		if _, err := parse(append(cliArgs(), mode)); err == nil {
			t.Fatal("missing writer readiness accepted", mode)
		}
		if _, err := parse(append(cliArgs(), mode, "-writers-ready")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := parse(append(cliArgs(), "-apply", "-rollback", "-writers-ready")); err == nil {
		t.Fatal("conflicting modes accepted")
	}
}

func TestScopedPathsAndLosslessConversion(t *testing.T) {
	rid, sid, pid := strings.Repeat("a", 32), "C"+strings.Repeat("b", 32), strings.Repeat("c", 32)
	prefix := "channels/123/repositories/" + rid + "/"
	target, root, err := targetPath("123", rid, sid, prefix+"projects/"+pid)
	if err != nil || !root || target != "channels/123/spaces/"+sid+"/projects/"+pid {
		t.Fatal(target, root, err)
	}
	target, root, err = targetPath("123", rid, sid, prefix+"projectsCreationReceipts/receipt")
	if err != nil || root || target != "channels/123/spaces/"+sid+"/projectCreationScopes/"+rid+"/receipts/receipt" {
		t.Fatal(target, root, err)
	}
	for _, source := range []string{"channels/999/repositories/" + rid + "/projects/" + pid, prefix + "issues/" + pid, prefix + "projects/../changes/x"} {
		if _, _, err := targetPath("123", rid, sid, source); err == nil {
			t.Fatal("out-of-scope path accepted", source)
		}
	}
	before := map[string]any{"repositoryId": rid, "id": pid, "version": int64(7), "createdAt": time.Unix(123, 456).UTC()}
	oldHash, _ := digest(before)
	after := converted(before, true, rid, sid)
	newHash, _ := digest(before)
	if oldHash != newHash || after["repositoryId"] != rid || after["version"] != int64(7) || after["spaceId"] != sid || after["sourceRepositoryId"] != rid {
		t.Fatal("conversion mutated source or changed existing metadata")
	}
	if _, ok := converted(before, false, rid, sid)["spaceId"]; ok {
		t.Fatal("child documents must remain unchanged")
	}
}

func TestManifestCannotChangeNamespaceOrDestination(t *testing.T) {
	o, _ := parse(cliArgs())
	path := filepath.Join(t.TempDir(), "manifest.json")
	o.Manifest = path
	sid := "C" + strings.Repeat("b", 32)
	source := "channels/123/repositories/" + o.Repository + "/projects/" + strings.Repeat("c", 32)
	target, _, _ := targetPath(o.Channel, o.Repository, sid, source)
	m := manifest{SchemaVersion: 1, MigrationID: strings.Repeat("d", 32), Project: o.Project, Database: o.Database, Channel: o.Channel, Repository: o.Repository, SpaceID: sid, Phase: "planned", Records: []record{{source, target, strings.Repeat("e", 64), strings.Repeat("f", 64)}}}
	if err := save(path, m); err != nil {
		t.Fatal(err)
	}
	if _, err := load(path, o); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatal("manifest must be private")
	}
	wrong := o
	wrong.Database = "other"
	if _, err := load(path, wrong); err == nil {
		t.Fatal("cross database manifest accepted")
	}
	m.Records[0].Target = "channels/123/other/arbitrary"
	if err := save(path, m); err != nil {
		t.Fatal(err)
	}
	if _, err := load(path, o); err == nil {
		t.Fatal("tampered target accepted")
	}
}

// Never opens production: this test is opt-in and requires a local emulator.
func TestMigrationLifecycle(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("requires FIRESTORE_EMULATOR_HOST; never use live credentials")
	}
	ctx := context.Background()
	client, err := db.NewClient(ctx, "project-migration-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	fixture := func(t *testing.T) (*migrator, manifest) {
		t.Helper()
		rid := strings.Repeat("a", 32)
		sid := "C" + strings.Repeat("b", 32)
		pid := strings.Repeat("c", 32)
		o := options{Project: "project-migration-tests", Database: "(default)", Channel: strconv.FormatInt(time.Now().UnixNano(), 10), Repository: rid, Manifest: filepath.Join(t.TempDir(), "manifest.json")}
		m := &migrator{client: client, options: o}
		now := time.Unix(1700000000, 123000).UTC()
		data := map[string]map[string]any{
			m.repository().Path: {"id": rid, "groupId": sid, "name": "Group", "status": "ACTIVE", "version": int64(1), "createdAt": now, "updatedAt": now, "eventAt": now, "eventId": "event"},
			m.repository().Collection("projects").Doc(pid).Path:                                 {"schemaVersion": int64(1), "id": pid, "repositoryId": rid, "title": "Project", "body": "Body", "creatorId": "member", "version": int64(7), "createdAt": now, "updatedAt": now},
			m.repository().Collection("projects").Doc(pid).Collection("items").Doc("item").Path: {"issueId": "issue", "projectId": pid, "createdAt": now},
			m.repository().Collection("projects").Doc(pid).Collection("changes").Doc("7").Path:  {"version": int64(7), "actor": "member"},
			m.repository().Collection("projectsCreationReceipts").Doc("receipt").Path:           {"id": pid, "payloadHash": "unchanged", "schemaVersion": int64(1)},
		}
		for path, d := range data {
			if _, err := client.Doc(relativePath(path)).Set(ctx, d); err != nil {
				t.Fatal(err)
			}
		}
		h := sha256.Sum256([]byte("member"))
		user := hex.EncodeToString(h[:])
		candidate := m.channel().Collection("repositoryUsers").Doc(user).Collection("repositories").Doc(rid)
		if _, err := candidate.Set(ctx, map[string]any{"repositoryId": rid}); err != nil {
			t.Fatal(err)
		}
		plan, err := m.plan(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.Records) != 4 {
			t.Fatalf("missing nested records: %d", len(plan.Records))
		}
		if marker, _ := m.marker().Get(ctx); marker.Exists() {
			t.Fatal("dry-run created marker")
		}
		if err := save(o.Manifest, plan); err != nil {
			t.Fatal(err)
		}
		return m, plan
	}
	t.Run("apply preserves data and rollback keeps both copies", func(t *testing.T) {
		m, plan := fixture(t)
		if err := m.apply(ctx, &plan); err != nil {
			t.Fatal(err)
		}
		if plan.Phase != "active" {
			t.Fatal(plan.Phase)
		}
		if err := m.verifySource(ctx, plan); err != nil {
			t.Fatal("legacy source changed", err)
		}
		if err := m.verifyTarget(ctx, plan); err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256([]byte("member"))
		candidate, err := m.channel().Collection("spaceUsers").Doc(hex.EncodeToString(h[:])).Collection("spaces").Doc(plan.SpaceID).Get(ctx)
		if err != nil || candidate.Data()["spaceId"] != plan.SpaceID {
			t.Fatal("migrated Space not discoverable", err)
		}
		plan.Phase = "locked"
		if err := m.apply(ctx, &plan); err != nil || plan.Phase != "active" {
			t.Fatal("cutover crash recovery failed", err)
		}
		if err := m.rollback(ctx, &plan); err != nil {
			t.Fatal(err)
		}
		if plan.Phase != "rolled_back" {
			t.Fatal(plan.Phase)
		}
		if err := m.verifyTarget(ctx, plan); err != nil {
			t.Fatal("rollback destroyed copy", err)
		}
	})
	t.Run("new canonical writes refuse rollback without losing data", func(t *testing.T) {
		m, plan := fixture(t)
		if err := m.apply(ctx, &plan); err != nil {
			t.Fatal(err)
		}
		target := m.client.Doc(plan.Records[0].Target)
		if _, err := target.Update(ctx, []db.Update{{Path: "version", Value: int64(99)}}); err != nil {
			t.Fatal(err)
		}
		if err := m.rollback(ctx, &plan); err == nil {
			t.Fatal("rollback accepted drift")
		}
		snap, err := m.marker().Get(ctx)
		if err != nil || snap.Data()["location"] != "space" || snap.Data()["locked"] != true {
			t.Fatal("unsafe marker after refusal", err)
		}
		if err := m.releaseLock(ctx, plan); err != nil {
			t.Fatal(err)
		}
		snap, _ = target.Get(ctx)
		if snap.Data()["version"] != int64(99) {
			t.Fatal("new write lost")
		}
	})
	t.Run("conflicting target cannot be overwritten", func(t *testing.T) {
		m, plan := fixture(t)
		if _, err := m.client.Doc(plan.Records[0].Target).Set(ctx, map[string]any{"existing": "keep"}); err != nil {
			t.Fatal(err)
		}
		if err := m.apply(ctx, &plan); err == nil {
			t.Fatal("target conflict accepted")
		}
		snap, _ := m.client.Doc(plan.Records[0].Target).Get(ctx)
		if snap.Data()["existing"] != "keep" {
			t.Fatal("target overwritten")
		}
	})
	t.Run("source drift and competing lock owners fail closed", func(t *testing.T) {
		m, plan := fixture(t)
		if err := m.lock(ctx, plan, "legacy"); err != nil {
			t.Fatal(err)
		}
		other := plan
		other.MigrationID = strings.Repeat("e", 32)
		if err := m.lock(ctx, other, "legacy"); err == nil {
			t.Fatal("competing lock accepted")
		}
		if err := m.releaseLock(ctx, other); err == nil {
			t.Fatal("foreign manifest released lock")
		}
		source := m.client.Doc(plan.Records[0].Source)
		if _, err := source.Update(ctx, []db.Update{{Path: "version", Value: int64(8)}}); err != nil {
			t.Fatal(err)
		}
		if err := m.apply(ctx, &plan); err == nil {
			t.Fatal("source drift accepted")
		}
		snap, _ := m.marker().Get(ctx)
		if snap.Data()["location"] != "legacy" || snap.Data()["locked"] != true {
			t.Fatal("unsafe source drift marker")
		}
	})
}
