package firestore

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"

	db "cloud.google.com/go/firestore"
	feature "github.com/963sup/line-mini-app-bot/internal/business/project"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type itemDocument struct {
	ItemStatus         string    `firestore:"itemStatus,omitempty"`
	SourceRepositoryID string    `firestore:"sourceRepositoryId,omitempty"`
	SchemaVersion      int64     `firestore:"schemaVersion"`
	ID                 string    `firestore:"id"`
	ProjectID          string    `firestore:"projectId"`
	SourceType         string    `firestore:"sourceType,omitempty"`
	SourceID           string    `firestore:"sourceId,omitempty"`
	ParentObjectID     string    `firestore:"parentObjectId,omitempty"`
	CreatedBy          string    `firestore:"createdBy"`
	CreatedAt          time.Time `firestore:"createdAt"`
}

func (s *Store) sourceIssue(tx *db.Transaction, source feature.Source, id string) (feature.Item, error) {
	if err := s.checkSourceRepository(tx, source); err != nil {
		return feature.Item{}, err
	}
	snap, e := tx.Get(s.repositoryRef(source.ID).Collection("issues").Doc(id))
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
	if d.SchemaVersion != 1 || d.ID != id || d.RepositoryID != source.ID || d.Title == "" || (d.State != "open" && d.State != "in_progress" && d.State != "completed") {
		return feature.Item{}, feature.ErrUnavailable
	}
	return feature.Item{IssueID: id, Title: d.Title, State: d.State}, nil
}

func decodeItem(snap *db.DocumentSnapshot, pid string) (itemDocument, error) {
	var d itemDocument
	if e := snap.DataTo(&d); e != nil {
		return itemDocument{}, e
	}
	ref := itemReference(d)
	if d.ID != snap.Ref.ID || !feature.ValidID(d.ID) || d.ProjectID != pid || d.CreatedBy == "" || d.CreatedAt.IsZero() || !feature.ValidReference(ref) || d.ID != referenceID(pid, ref) {
		return itemDocument{}, feature.ErrUnavailable
	}
	if d.SchemaVersion != 2 || !feature.ValidID(d.SourceRepositoryID) {
		return itemDocument{}, feature.ErrUnavailable
	}
	return d, nil
}

func (s *Store) readItem(tx *db.Transaction, r feature.Scope, pid string, snap *db.DocumentSnapshot) (feature.Item, error) {
	d, e := decodeItem(snap, pid)
	if e != nil {
		return feature.Item{}, e
	}
	if d.SourceRepositoryID != r.SourceRepositoryID || itemReference(d).SourceType != "issue" {
		return feature.Item{}, feature.ErrInput
	}
	i, e := s.sourceIssue(tx, feature.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}, d.SourceID)
	if e != nil {
		return feature.Item{}, e
	}
	i.ID = d.ID
	return i, nil
}

func requireProjectLink(tx *db.Transaction, project *db.DocumentRef, sourceID string) error {
	_, err := tx.Get(project.Collection("linkedRepositories").Doc(sourceID))
	if status.Code(err) == codes.NotFound {
		return feature.ErrSourceUnavailable
	}
	return err
}

func (s *Store) Progress(ctx context.Context, r feature.Scope, result *feature.Project) (err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	r.SourceRepositoryID = result.DefaultRepositoryID
	parent := s.ownerDoc(r.OwnerID)
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.checkProjectOwner(tx, r, parent); e != nil {
			return e
		}
		if e := requireProjectLink(tx, parent.Collection("projects").Doc(result.ID), r.SourceRepositoryID); e != nil {
			return e
		}
		result.IssueCount = 0
		result.CompletedIssueCount = 0
		snaps, e := tx.Documents(parent.Collection("projects").Doc(result.ID).Collection("items")).GetAll()
		if e != nil {
			return e
		}
		for _, snap := range snaps {
			d, e := decodeItem(snap, result.ID)
			if e != nil {
				return e
			}
			if d.SourceRepositoryID != r.SourceRepositoryID || itemReference(d).SourceType != "issue" {
				continue
			}
			i, e := s.sourceIssue(tx, feature.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}, d.SourceID)
			if e != nil {
				if r.Legacy || errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) || status.Code(e) == codes.Canceled || status.Code(e) == codes.DeadlineExceeded {
					return e
				}
				return feature.ErrSourceUnavailable
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
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	result.Items = []feature.Item{}
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	scope := []string{s.ChannelID, r.OwnerID, actor, pid}
	if r.Legacy {
		scope[1] = r.SourceRepositoryID
	}
	root := s.ownerDoc(r.OwnerID)
	parent := root.Collection("projects").Doc(pid)
	q, e := applyListCursor(parent.Collection("items").OrderBy(db.DocumentID, db.Asc), token, scope)
	if e != nil {
		return result, e
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.ItemPage{Items: []feature.Item{}}
		if e := s.checkProjectOwner(tx, r, root); e != nil {
			return e
		}
		snap, e := tx.Get(parent)
		if _, e := decodeProject(snap, e, r); e != nil {
			return e
		}
		if e := requireProjectLink(tx, parent, r.SourceRepositoryID); e != nil {
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
			d, e := decodeItem(snap, pid)
			if e != nil {
				return e
			}
			if d.SourceRepositoryID != r.SourceRepositoryID || itemReference(d).SourceType != "issue" {
				continue
			}
			i, e := s.readItem(tx, r, pid, snap)
			if e != nil {
				return e
			}
			result.Items = append(result.Items, i)
		}
		if more {
			result.NextCursor = firestorekey.EncodeCursor(scope, snaps[len(snaps)-1].Ref.ID)
		}
		return nil
	}, db.ReadOnly)
	if err != nil {
		return feature.ItemPage{}, err
	}
	return result, nil
}

func (s *Store) ChangeItem(ctx context.Context, r feature.Scope, actor, pid, target string, add bool, version int64) (result feature.Project, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	root := s.ownerDoc(r.OwnerID)
	parent := root.Collection("projects").Doc(pid)
	itemID := target
	if add {
		itemID = referenceID(pid, feature.Reference{SourceRepositoryID: r.SourceRepositoryID, SourceID: target, SourceType: "issue"})
	}
	ref := parent.Collection("items").Doc(itemID)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.checkProjectOwner(tx, r, root); e != nil {
			return e
		}
		snap, e := tx.Get(parent)
		current, e := decodeProject(snap, e, r)
		if e != nil {
			return e
		}
		if e := feature.CheckUpdate(current, actor, version); e != nil {
			return e
		}
		if e := requireProjectLink(tx, parent, r.SourceRepositoryID); e != nil {
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
			if _, e := s.sourceIssue(tx, feature.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}, target); e != nil {
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
			if e := tx.Create(ref, itemDocument{SchemaVersion: 2, ID: itemID, ProjectID: pid, SourceRepositoryID: r.SourceRepositoryID, SourceType: "issue", SourceID: target, CreatedBy: actor, CreatedAt: now}); e != nil {
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
		return writeProjectChange(tx, parent, current, updated, actor, action)
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, pid)
}
