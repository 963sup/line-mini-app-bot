package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	db "cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type migrator struct {
	client  *db.Client
	options options
}
type storageMarker struct {
	SpaceID     string `firestore:"spaceId"`
	Location    string `firestore:"location"`
	Locked      bool   `firestore:"locked"`
	MigrationID string `firestore:"migrationId"`
}

func (m *migrator) channel() *db.DocumentRef {
	return m.client.Collection("channels").Doc(m.options.Channel)
}

func (m *migrator) repository() *db.DocumentRef {
	return m.channel().Collection("repositories").Doc(m.options.Repository)
}

func (m *migrator) marker() *db.DocumentRef {
	return m.channel().Collection("projectStorage").Doc(m.options.Repository)
}

func (m *migrator) space(id string) *db.DocumentRef { return m.channel().Collection("spaces").Doc(id) }

func targetPath(channel, rid, sid, source string) (string, bool, error) {
	root := "channels/" + channel + "/repositories/" + rid + "/"
	if !strings.HasPrefix(source, root) {
		return "", false, errors.New("source is outside the explicit repository")
	}
	tail := strings.TrimPrefix(source, root)
	parts := strings.Split(tail, "/")
	if len(parts) < 2 || len(parts)%2 != 0 {
		return "", false, errors.New("invalid source document path")
	}
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return "", false, errors.New("invalid path segment")
		}
	}
	switch parts[0] {
	case "projects":
		if !idPattern.MatchString(parts[1]) {
			return "", false, errors.New("invalid project ID")
		}
		return "channels/" + channel + "/spaces/" + sid + "/" + tail, len(parts) == 2, nil
	case "projectsCreationReceipts":
		if len(parts) != 2 {
			return "", false, errors.New("unexpected nested creation receipt")
		}
		return "channels/" + channel + "/spaces/" + sid + "/projectCreationScopes/" + rid + "/receipts/" + parts[1], false, nil
	default:
		return "", false, errors.New("unsupported migration source")
	}
}

func converted(data map[string]any, root bool, rid, sid string) map[string]any {
	result := make(map[string]any, len(data)+2)
	for k, v := range data {
		result[k] = v
	}
	if root {
		result["spaceId"] = sid
		result["sourceRepositoryId"] = rid
	}
	return result
}

// DocumentRefs includes missing parents with child collections. Refuse orphans
// rather than silently dropping their items or audit records.
func collect(ctx context.Context, collection *db.CollectionRef, out *[]*db.DocumentSnapshot) error {
	refs := collection.DocumentRefs(ctx)
	for {
		ref, err := refs.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}
		snap, err := ref.Get(ctx)
		if err != nil {
			return errors.New("missing or unreadable document in migration subtree")
		}
		*out = append(*out, snap)
		children := ref.Collections(ctx)
		for {
			child, err := children.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			if err := collect(ctx, child, out); err != nil {
				return err
			}
		}
	}
}

func (m *migrator) source(ctx context.Context) ([]*db.DocumentSnapshot, error) {
	var docs []*db.DocumentSnapshot
	for _, kind := range []string{"projects", "projectsCreationReceipts"} {
		if err := collect(ctx, m.repository().Collection(kind), &docs); err != nil {
			return nil, err
		}
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Ref.Path < docs[j].Ref.Path })
	return docs, nil
}

func (m *migrator) repositoryData(ctx context.Context) (map[string]any, string, error) {
	snap, err := m.repository().Get(ctx)
	if err != nil {
		return nil, "", errors.New("explicit source repository is missing or unreadable")
	}
	d := snap.Data()
	id, _ := d["id"].(string)
	group, _ := d["groupId"].(string)
	if id != m.options.Repository || !groupPattern.MatchString(group) {
		return nil, "", errors.New("invalid repository/group mapping")
	}
	return d, group, nil
}

func (m *migrator) plan(ctx context.Context) (manifest, error) {
	o := m.options
	plan := manifest{SchemaVersion: 1, Project: o.Project, Database: o.Database, Channel: o.Channel, Repository: o.Repository, Phase: "planned"}
	_, sid, err := m.repositoryData(ctx)
	if err != nil {
		return plan, err
	}
	plan.SpaceID = sid
	plan.MigrationID, err = newID()
	if err != nil {
		return plan, err
	}
	marker, err := m.marker().Get(ctx)
	if err == nil {
		var s storageMarker
		if marker.DataTo(&s) != nil || s.SpaceID != sid || s.Location != "legacy" || s.Locked {
			return plan, errors.New("source is already switched, locked or has a conflicting marker")
		}
	} else if status.Code(err) != codes.NotFound {
		return plan, err
	}
	docs, err := m.source(ctx)
	if err != nil {
		return plan, err
	}
	for _, snap := range docs {
		source := relativePath(snap.Ref.Path)
		target, root, err := targetPath(o.Channel, o.Repository, sid, source)
		if err != nil {
			return plan, err
		}
		data := snap.Data()
		if root && (data["id"] != snap.Ref.ID || data["repositoryId"] != o.Repository) {
			return plan, errors.New("legacy project identity mismatch")
		}
		a, err := digest(data)
		if err != nil {
			return plan, err
		}
		b, err := digest(converted(data, root, o.Repository, sid))
		if err != nil {
			return plan, err
		}
		plan.Records = append(plan.Records, record{source, target, a, b})
		existing, err := m.client.Doc(target).Get(ctx)
		if err == nil {
			h, e := digest(existing.Data())
			if e != nil || h != b {
				return plan, errors.New("target conflict; no document will be overwritten")
			}
		} else if status.Code(err) != codes.NotFound {
			return plan, err
		}
	}
	return plan, nil
}

func relativePath(path string) string {
	if _, tail, ok := strings.Cut(path, "/documents/"); ok {
		return tail
	}
	return path
}

func (m *migrator) lock(ctx context.Context, plan manifest, location string) error {
	return m.client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		snap, err := tx.Get(m.marker())
		state := storageMarker{SpaceID: plan.SpaceID, Location: "legacy"}
		if err == nil {
			if snap.DataTo(&state) != nil {
				return errors.New("invalid storage marker")
			}
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		if state.SpaceID != plan.SpaceID || state.Location != location || (state.Locked && state.MigrationID != plan.MigrationID) {
			return errors.New("marker location, identity or lock owner changed")
		}
		return tx.Set(m.marker(), map[string]any{"spaceId": plan.SpaceID, "location": location, "locked": true, "migrationId": plan.MigrationID})
	})
}

func (m *migrator) checkLock(tx *db.Transaction, plan manifest, location string) error {
	snap, err := tx.Get(m.marker())
	if err != nil {
		return err
	}
	var s storageMarker
	if snap.DataTo(&s) != nil || s.SpaceID != plan.SpaceID || s.Location != location || !s.Locked || s.MigrationID != plan.MigrationID {
		return errors.New("migration lock no longer held")
	}
	return nil
}

func (m *migrator) switchLocation(ctx context.Context, plan manifest, from, to string) error {
	return m.client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if err := m.checkLock(tx, plan, from); err != nil {
			return err
		}
		return tx.Update(m.marker(), []db.Update{{Path: "location", Value: to}, {Path: "locked", Value: false}})
	})
}

func (m *migrator) verifySource(ctx context.Context, plan manifest) error {
	_, sid, err := m.repositoryData(ctx)
	if err != nil {
		return err
	}
	if sid != plan.SpaceID {
		return errors.New("source group mapping changed")
	}
	docs, err := m.source(ctx)
	if err != nil {
		return err
	}
	if len(docs) != len(plan.Records) {
		return errors.New("source document set changed; migration remains locked")
	}
	for i, snap := range docs {
		h, err := digest(snap.Data())
		if err != nil {
			return err
		}
		r := plan.Records[i]
		if relativePath(snap.Ref.Path) != r.Source || h != r.SourceHash {
			return errors.New("source data changed; migration remains locked")
		}
	}
	return nil
}

func (m *migrator) bootstrap(ctx context.Context, plan manifest) error {
	d, group, err := m.repositoryData(ctx)
	if err != nil {
		return err
	}
	if group != plan.SpaceID {
		return errors.New("source group mapping changed")
	}
	spaceData := map[string]any{"id": group, "groupId": group}
	for _, k := range []string{"name", "status", "version", "createdAt", "updatedAt", "eventAt", "eventId"} {
		spaceData[k] = d[k]
	}
	if err := m.client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if err := m.checkLock(tx, plan, "legacy"); err != nil {
			return err
		}
		snap, err := tx.Get(m.space(group))
		if err == nil {
			v := snap.Data()
			if v["id"] != group || v["groupId"] != group {
				return errors.New("space identity collision")
			}
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		return tx.Create(m.space(group), spaceData)
	}); err != nil {
		return err
	}
	users := m.channel().Collection("repositoryUsers").DocumentRefs(ctx)
	for {
		user, err := users.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}
		candidate := user.Collection("repositories").Doc(plan.Repository)
		snap, err := candidate.Get(ctx)
		if status.Code(err) == codes.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		if snap.Data()["repositoryId"] != plan.Repository {
			return errors.New("invalid legacy candidate")
		}
		target := m.channel().Collection("spaceUsers").Doc(user.ID).Collection("spaces").Doc(group)
		if err := m.client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
			if err := m.checkLock(tx, plan, "legacy"); err != nil {
				return err
			}
			existing, err := tx.Get(target)
			if err == nil {
				if existing.Data()["spaceId"] != group {
					return errors.New("space candidate collision")
				}
				return nil
			}
			if status.Code(err) != codes.NotFound {
				return err
			}
			return tx.Create(target, map[string]any{"spaceId": group})
		}); err != nil {
			return err
		}
	}
}

func (m *migrator) copyRecord(ctx context.Context, plan manifest, r record) error {
	return m.client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if err := m.checkLock(tx, plan, "legacy"); err != nil {
			return err
		}
		source, err := tx.Get(m.client.Doc(r.Source))
		if err != nil {
			return err
		}
		h, err := digest(source.Data())
		if err != nil || h != r.SourceHash {
			return errors.New("source changed during copy")
		}
		_, root, err := targetPath(plan.Channel, plan.Repository, plan.SpaceID, r.Source)
		if err != nil {
			return err
		}
		data := converted(source.Data(), root, plan.Repository, plan.SpaceID)
		h, err = digest(data)
		if err != nil || h != r.TargetHash {
			return errors.New("manifest target hash does not match conversion")
		}
		target, err := tx.Get(m.client.Doc(r.Target))
		if err == nil {
			h, e := digest(target.Data())
			if e != nil || h != r.TargetHash {
				return errors.New("target conflict; refusing overwrite")
			}
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		return tx.Create(m.client.Doc(r.Target), data)
	})
}

func (m *migrator) verifyTarget(ctx context.Context, plan manifest) error {
	expected := map[string]string{}
	roots := map[string]bool{}
	for _, r := range plan.Records {
		expected[r.Target] = r.TargetHash
		_, root, _ := targetPath(plan.Channel, plan.Repository, plan.SpaceID, r.Source)
		if root {
			roots[r.Target] = true
		}
	}
	actual := map[string]string{}
	for path := range roots {
		ref := m.client.Doc(path)
		snap, err := ref.Get(ctx)
		if err != nil {
			return errors.New("target root missing")
		}
		h, err := digest(snap.Data())
		if err != nil {
			return err
		}
		actual[path] = h
		children := ref.Collections(ctx)
		for {
			child, err := children.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			var docs []*db.DocumentSnapshot
			if err := collect(ctx, child, &docs); err != nil {
				return err
			}
			for _, d := range docs {
				h, err := digest(d.Data())
				if err != nil {
					return err
				}
				actual[relativePath(d.Ref.Path)] = h
			}
		}
	}
	var receipts []*db.DocumentSnapshot
	if err := collect(ctx, m.space(plan.SpaceID).Collection("projectCreationScopes").Doc(plan.Repository).Collection("receipts"), &receipts); err != nil {
		return err
	}
	for _, snap := range receipts {
		h, err := digest(snap.Data())
		if err != nil {
			return err
		}
		actual[relativePath(snap.Ref.Path)] = h
	}
	// Detect newly created projects with this source, including roots not in the plan.
	projects, err := m.space(plan.SpaceID).Collection("projects").Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	for _, snap := range projects {
		if snap.Data()["sourceRepositoryId"] == plan.Repository {
			path := relativePath(snap.Ref.Path)
			if !roots[path] {
				return errors.New("new canonical project detected; rollback/cutover refused")
			}
		}
	}
	if len(actual) != len(expected) {
		return errors.New("target subtree changed; migration remains locked")
	}
	for path, h := range expected {
		if actual[path] != h {
			return errors.New("target data changed; migration remains locked")
		}
	}
	return nil
}

func (m *migrator) apply(ctx context.Context, plan *manifest) error {
	// Recover a successful cutover whose local manifest write was interrupted.
	if plan.Phase == "locked" {
		recovered, err := m.completedSwitch(ctx, *plan, "space")
		if err != nil {
			return err
		}
		if recovered {
			plan.Phase = "active"
			return save(m.options.Manifest, *plan)
		}
	}
	if plan.Phase == "active" {
		return m.verifyTarget(ctx, *plan)
	}
	if plan.Phase != "planned" && plan.Phase != "locked" {
		return errors.New("manifest cannot be applied in this phase")
	}
	if err := m.lock(ctx, *plan, "legacy"); err != nil {
		return err
	}
	plan.Phase = "locked"
	if err := save(m.options.Manifest, *plan); err != nil {
		return err
	}
	if err := m.verifySource(ctx, *plan); err != nil {
		return err
	}
	if err := m.bootstrap(ctx, *plan); err != nil {
		return err
	}
	for _, r := range plan.Records {
		if err := m.copyRecord(ctx, *plan, r); err != nil {
			return err
		}
	}
	if err := m.verifySource(ctx, *plan); err != nil {
		return err
	}
	if err := m.verifyTarget(ctx, *plan); err != nil {
		return err
	}
	if err := m.switchLocation(ctx, *plan, "legacy", "space"); err != nil {
		return err
	}
	plan.Phase = "active"
	return save(m.options.Manifest, *plan)
}

func (m *migrator) rollback(ctx context.Context, plan *manifest) error {
	// As above, the atomic marker is authoritative if local persistence failed.
	if plan.Phase == "active" {
		recovered, err := m.completedSwitch(ctx, *plan, "legacy")
		if err != nil {
			return err
		}
		if recovered {
			plan.Phase = "rolled_back"
			return save(m.options.Manifest, *plan)
		}
	}
	if plan.Phase != "active" {
		return errors.New("rollback requires the completed apply manifest")
	}
	if err := m.lock(ctx, *plan, "space"); err != nil {
		return err
	}
	if err := m.verifySource(ctx, *plan); err != nil {
		return fmt.Errorf("rollback refused: %w", err)
	}
	if err := m.verifyTarget(ctx, *plan); err != nil {
		return fmt.Errorf("rollback refused: %w", err)
	}
	if err := m.switchLocation(ctx, *plan, "space", "legacy"); err != nil {
		return err
	}
	plan.Phase = "rolled_back"
	return save(m.options.Manifest, *plan)
}

func (m *migrator) completedSwitch(ctx context.Context, plan manifest, location string) (bool, error) {
	snap, err := m.marker().Get(ctx)
	if status.Code(err) == codes.NotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var s storageMarker
	if snap.DataTo(&s) != nil {
		return false, errors.New("invalid storage marker")
	}
	return s.SpaceID == plan.SpaceID && s.MigrationID == plan.MigrationID && s.Location == location && !s.Locked, nil
}

func (m *migrator) releaseLock(ctx context.Context, plan manifest) error {
	return m.client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		snap, err := tx.Get(m.marker())
		if err != nil {
			return err
		}
		var s storageMarker
		if snap.DataTo(&s) != nil || s.SpaceID != plan.SpaceID || s.MigrationID != plan.MigrationID || !s.Locked || (s.Location != "legacy" && s.Location != "space") {
			return errors.New("this manifest does not own the lock")
		}
		return tx.Update(m.marker(), []db.Update{{Path: "locked", Value: false}})
	})
}
