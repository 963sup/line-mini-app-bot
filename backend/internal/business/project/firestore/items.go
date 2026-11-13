package firestore

import (
	"context"
	"time"

	db "cloud.google.com/go/firestore"
	feature "github.com/963sup/line-mini-app-bot/internal/business/project"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type itemDocument struct {
	SchemaVersion int64     `firestore:"schemaVersion"`
	ID            string    `firestore:"id"`
	ProjectID     string    `firestore:"projectId"`
	IssueID       string    `firestore:"issueId"`
	CreatedBy     string    `firestore:"createdBy"`
	CreatedAt     time.Time `firestore:"createdAt"`
}

func (s *Store) sourceIssue(tx *db.Transaction, r feature.Scope, id string) (feature.Item, error) {
	source, e := tx.Get(s.legacyParent(r.SourceRepositoryID))
	if e != nil {
		return feature.Item{}, e
	}
	var container struct{ ID, GroupID, Status string }
	if e = source.DataTo(&container); e != nil {
		return feature.Item{}, e
	}
	if container.ID != r.SourceRepositoryID || container.GroupID != r.SpaceID || container.Status != "ACTIVE" {
		return feature.Item{}, feature.ErrNotFound
	}
	snap, e := tx.Get(s.legacyParent(r.SourceRepositoryID).Collection("issues").Doc(id))
	if status.Code(e) == codes.NotFound {
		return feature.Item{}, feature.ErrInput
	}
	if e != nil {
		return feature.Item{}, e
	}
	var d struct {
		SchemaVersion int64  `firestore:"schemaVersion"`
		ID            string `firestore:"id"`
		RepositoryID  string `firestore:"repositoryId"`
		Title         string `firestore:"title"`
		State         string `firestore:"state"`
	}
	if e := snap.DataTo(&d); e != nil {
		return feature.Item{}, e
	}
	if d.SchemaVersion != 1 || d.ID != id || d.RepositoryID != r.SourceRepositoryID || d.Title == "" || (d.State != "open" && d.State != "in_progress" && d.State != "completed") {
		return feature.Item{}, feature.ErrUnavailable
	}
	return feature.Item{IssueID: id, Title: d.Title, State: d.State}, nil
}

func (s *Store) readItem(tx *db.Transaction, r feature.Scope, pid string, snap *db.DocumentSnapshot) (feature.Item, error) {
	var d itemDocument
	if e := snap.DataTo(&d); e != nil {
		return feature.Item{}, e
	}
	if d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !feature.ValidID(d.ID) || d.ProjectID != pid || !feature.ValidID(d.IssueID) || d.CreatedBy == "" || d.CreatedAt.IsZero() || d.ID != planningHash([]string{"project-item", pid, d.IssueID})[:32] {
		return feature.Item{}, feature.ErrUnavailable
	}
	i, e := s.sourceIssue(tx, r, d.IssueID)
	if e != nil {
		return feature.Item{}, e
	}
	i.ID = d.ID
	return i, nil
}

func (s *Store) Progress(ctx context.Context, r feature.Scope, result *feature.Project) (err error) {
	defer func() { err = planningError(err) }()
	r.SourceRepositoryID = result.SourceRepositoryID
	parent, e := s.readParent(ctx, r)
	if e != nil {
		return e
	}
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.planningContainer(tx, r, parent); e != nil {
			return e
		}
		result.IssueCount = 0
		result.CompletedIssueCount = 0
		snaps, e := tx.Documents(parent.Collection("projects").Doc(result.ID).Collection("items")).GetAll()
		if e != nil {
			return e
		}
		for _, snap := range snaps {
			i, e := s.readItem(tx, r, result.ID, snap)
			if e != nil {
				return e
			}
			result.IssueCount++
			if i.State == "completed" {
				result.CompletedIssueCount++
			}
		}
		return nil
	}, db.ReadOnly)
}

func (s *Store) ListItems(ctx context.Context, r feature.Scope, actor, pid, token string) (result feature.ItemPage, err error) {
	defer func() { err = planningError(err) }()
	result.Items = []feature.Item{}
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	scope := []string{s.ChannelID, r.SpaceID, actor, pid}
	if r.Legacy {
		scope[1] = r.SourceRepositoryID
	}
	root, e := s.readParent(ctx, r)
	if e != nil {
		return result, e
	}
	parent := root.Collection("projects").Doc(pid)
	q, e := planningPage(parent.Collection("items").OrderBy(db.DocumentID, db.Asc), token, scope)
	if e != nil {
		return result, e
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.ItemPage{Items: []feature.Item{}}
		if e := s.planningContainer(tx, r, root); e != nil {
			return e
		}
		snap, e := tx.Get(parent)
		if _, e := planningDecode(snap, e, r); e != nil {
			return e
		}
		snaps, e := tx.Documents(q).GetAll()
		if e != nil {
			return e
		}
		more := len(snaps) > 20
		if more {
			snaps = snaps[:20]
		}
		for _, snap := range snaps {
			i, e := s.readItem(tx, r, pid, snap)
			if e != nil {
				return e
			}
			result.Items = append(result.Items, i)
		}
		if more {
			result.NextCursor = planningNext(scope, snaps[len(snaps)-1].Ref.ID)
		}
		return nil
	}, db.ReadOnly)
	if err != nil {
		return feature.ItemPage{}, err
	}
	return result, nil
}

func (s *Store) ChangeItem(ctx context.Context, r feature.Scope, actor, pid, target string, add bool, version int64) (result feature.Project, err error) {
	defer func() { err = planningError(err) }()
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	root, e := s.readParent(ctx, r)
	if e != nil {
		return result, e
	}
	parent := root.Collection("projects").Doc(pid)
	itemID := target
	if add {
		itemID = planningHash([]string{"project-item", pid, target})[:32]
	}
	ref := parent.Collection("items").Doc(itemID)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.planningContainer(tx, r, root); e != nil {
			return e
		}
		snap, e := tx.Get(parent)
		current, e := planningDecode(snap, e, r)
		if e != nil {
			return e
		}
		if e := feature.CheckUpdate(current, actor, version); e != nil {
			return e
		}
		item, e := tx.Get(ref)
		exists := e == nil
		if e != nil && status.Code(e) != codes.NotFound {
			return e
		}
		if exists {
			if _, e := s.readItem(tx, r, pid, item); e != nil {
				return e
			}
		}
		if add {
			if _, e := s.sourceIssue(tx, r, target); e != nil {
				return e
			}
		}
		if exists == add {
			return nil
		}
		if current.Version >= feature.MaxVersion {
			return feature.ErrConflict
		}
		updated := current
		updated.Version++
		if add {
			if e := tx.Create(ref, itemDocument{1, itemID, pid, target, actor, now}); e != nil {
				return e
			}
		} else {
			if e := tx.Delete(ref); e != nil {
				return e
			}
		}
		action := "remove_item"
		if add {
			action = "add_item"
		}
		return s.planningWrite(tx, parent, current, updated, actor, action)
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, pid)
}
