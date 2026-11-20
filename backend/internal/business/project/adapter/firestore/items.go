package firestore

import (
	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"github.com/963sup/line-mini-app-bot/internal/platform/grpcstatus"

	db "cloud.google.com/go/firestore"

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

func (s *Store) sourceIssue(tx *db.Transaction, source projectDomainEntity.Source, id string) (projectDomainEntity.Item, error) {
	if err := s.checkSourceRepository(tx, source); err != nil {
		return projectDomainEntity.Item{}, err
	}
	snap, e := tx.Get(s.repositoryRef(source.ID).Collection("issues").Doc(id))
	if status.Code(e) == codes.NotFound {
		return projectDomainEntity.Item{}, projectDomainEntity.ErrInput
	}
	if e != nil {
		return projectDomainEntity.Item{}, e
	}
	var d struct {
		SchemaVersion int64  `firestore:"schemaVersion"`
		ID            string `firestore:"id"`
		RepositoryID  string `firestore:"repositoryId"`
		Title         string `firestore:"title"`
		State         string `firestore:"state"`
	}
	if e := snap.DataTo(&d); e != nil {
		return projectDomainEntity.Item{}, e
	}
	if d.SchemaVersion != 1 || d.ID != id || d.RepositoryID != source.ID || d.Title == "" || (d.State != "open" && d.State != "in_progress" && d.State != "completed") {
		return projectDomainEntity.Item{}, projectDomainEntity.ErrUnavailable
	}
	return projectDomainEntity.Item{IssueID: id, Title: d.Title, State: d.State}, nil
}

func decodeItem(snap *db.DocumentSnapshot, pid string) (itemDocument, error) {
	var d itemDocument
	if e := snap.DataTo(&d); e != nil {
		return itemDocument{}, e
	}
	ref := itemReference(d)
	if d.ID != snap.Ref.ID || !projectDomainEntity.ValidID(d.ID) || d.ProjectID != pid || d.CreatedBy == "" || d.CreatedAt.IsZero() || !projectDomainEntity.ValidReference(ref) || d.ID != referenceID(pid, ref) {
		return itemDocument{}, projectDomainEntity.ErrUnavailable
	}
	if d.SchemaVersion != 2 || !projectDomainEntity.ValidID(d.SourceRepositoryID) {
		return itemDocument{}, projectDomainEntity.ErrUnavailable
	}
	return d, nil
}

func (s *Store) readItem(tx *db.Transaction, r projectApplicationPort.Scope, pid string, snap *db.DocumentSnapshot) (projectDomainEntity.Item, error) {
	d, e := decodeItem(snap, pid)
	if e != nil {
		return projectDomainEntity.Item{}, e
	}
	if d.SourceRepositoryID != r.SourceRepositoryID || itemReference(d).SourceType != "issue" {
		return projectDomainEntity.Item{}, projectDomainEntity.ErrInput
	}
	i, e := s.sourceIssue(tx, projectDomainEntity.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}, d.SourceID)
	if e != nil {
		return projectDomainEntity.Item{}, e
	}
	i.ID = d.ID
	return i, nil
}

func requireProjectLink(tx *db.Transaction, project *db.DocumentRef, sourceID string) error {
	_, err := tx.Get(project.Collection("linkedRepositories").Doc(sourceID))
	if status.Code(err) == codes.NotFound {
		return projectDomainEntity.ErrSourceUnavailable
	}
	return err
}

func (s *Store) Progress(ctx context.Context, r projectApplicationPort.Scope, result *projectDomainEntity.Project) (err error) {
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
			i, e := s.sourceIssue(tx, projectDomainEntity.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}, d.SourceID)
			if e != nil {
				if r.Legacy || errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) || status.Code(e) == codes.Canceled || status.Code(e) == codes.DeadlineExceeded {
					return e
				}
				return projectDomainEntity.ErrSourceUnavailable
			}
			result.IssueCount++
			if i.State == "completed" {
				result.CompletedIssueCount++
			}
		}
		return nil
	}, db.ReadOnly)
}

func (s *Store) ListItems(ctx context.Context, r projectApplicationPort.Scope, actor, pid, token string) (result projectDomainEntity.ItemPage, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	result.Items = []projectDomainEntity.Item{}
	if !s.ready() {
		return result, projectDomainEntity.ErrUnavailable
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
		result = projectDomainEntity.ItemPage{Items: []projectDomainEntity.Item{}}
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
		return projectDomainEntity.ItemPage{}, err
	}
	return result, nil
}

func (s *Store) ChangeItem(ctx context.Context, r projectApplicationPort.Scope, actor, pid, target string, add bool, version int64) (result projectDomainEntity.Project, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, projectDomainEntity.ErrUnavailable
	}
	root := s.ownerDoc(r.OwnerID)
	parent := root.Collection("projects").Doc(pid)
	itemID := target
	if add {
		itemID = referenceID(pid, projectDomainEntity.Reference{SourceRepositoryID: r.SourceRepositoryID, SourceID: target, SourceType: "issue"})
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
		if e := projectDomainEntity.CheckUpdate(current, actor, version); e != nil {
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
			if _, e := s.sourceIssue(tx, projectDomainEntity.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}, target); e != nil {
				return e
			}
		}
		if exists == add {
			return nil
		}
		if current.Version >= projectDomainEntity.MaxVersion {
			return projectDomainEntity.ErrConflict
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
