package firestore

import (
	"context"
	"errors"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/object"
	feature "github.com/963sup/line-mini-app-bot/internal/business/project"
	child "github.com/963sup/line-mini-app-bot/internal/business/subobject"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
)

func itemReference(d itemDocument) feature.Reference {
	return feature.Reference{ID: d.ID, SourceRepositoryID: d.SourceRepositoryID, SourceType: d.SourceType, SourceID: d.SourceID, ParentObjectID: d.ParentObjectID}
}

func referenceID(pid string, r feature.Reference) string {
	return firestorekey.JSONDigest([]string{"project-repository-reference", pid, r.SourceRepositoryID, r.SourceType, r.ParentObjectID, r.SourceID})[:32]
}

func (s *Store) checkSourceRepository(tx *db.Transaction, source feature.Source) error {
	repository, e := tx.Get(s.repositoryRef(source.ID))
	if e != nil {
		return e
	}
	var repo struct{ ID, OwnerID, Status string }
	if e = repository.DataTo(&repo); e != nil {
		return e
	}
	if repo.ID != source.ID || repo.OwnerID != source.OwnerID || repo.Status != "ACTIVE" {
		return feature.ErrNotFound
	}
	return nil
}

func (s *Store) sourceReference(tx *db.Transaction, source feature.Source, ref feature.Reference) (feature.Reference, error) {
	ref.SourceRepositoryID = source.ID
	root := s.repositoryRef(source.ID)
	if ref.SourceType == "issue" {
		i, e := s.sourceIssue(tx, source, ref.SourceID)
		if e != nil {
			return feature.Reference{}, e
		}
		ref.SourceTitle = i.Title
		ref.SourceState = i.State
		return ref, nil
	}
	if err := s.checkSourceRepository(tx, source); err != nil {
		return feature.Reference{}, err
	}
	if ref.SourceType == "discussion" {
		snap, e := tx.Get(root.Collection("discussions").Doc(ref.SourceID))
		if e != nil {
			return feature.Reference{}, e
		}
		var d struct {
			SchemaVersion int64  `firestore:"schemaVersion"`
			ID            string `firestore:"id"`
			RepositoryID  string `firestore:"repositoryId"`
			Title         string `firestore:"title"`
		}
		if e = snap.DataTo(&d); e != nil {
			return feature.Reference{}, e
		}
		if d.SchemaVersion != 1 || d.ID != ref.SourceID || d.RepositoryID != source.ID || d.Title == "" {
			return feature.Reference{}, feature.ErrSourceUnavailable
		}
		ref.SourceTitle = d.Title
		return ref, nil
	}
	collection := root.Collection("objects")
	if ref.SourceType == "subobject" {
		parent, e := tx.Get(collection.Doc(ref.ParentObjectID))
		if e != nil {
			return feature.Reference{}, e
		}
		var p struct {
			ID            string `firestore:"id"`
			RepositoryID  string `firestore:"repositoryId"`
			SchemaVersion int64  `firestore:"schemaVersion"`
		}
		if e = parent.DataTo(&p); e != nil {
			return feature.Reference{}, e
		}
		if p.ID != ref.ParentObjectID || p.RepositoryID != source.ID || p.SchemaVersion != 1 {
			return feature.Reference{}, feature.ErrSourceUnavailable
		}
		collection = collection.Doc(ref.ParentObjectID).Collection("subobjects")
	}
	snap, e := tx.Get(collection.Doc(ref.SourceID))
	if e != nil {
		return feature.Reference{}, e
	}
	var d struct {
		SchemaVersion     int64  `firestore:"schemaVersion"`
		ID                string `firestore:"id"`
		RepositoryID      string `firestore:"repositoryId"`
		ParentObjectID    string `firestore:"parentObjectId"`
		Name              string `firestore:"name"`
		Type              string `firestore:"type"`
		Quantity          string `firestore:"quantity"`
		Unit              string `firestore:"unit"`
		Version           int64  `firestore:"version"`
		CompletedQuantity string `firestore:"completedQuantity"`
		Status            string `firestore:"status"`
	}
	if e = snap.DataTo(&d); e != nil {
		return feature.Reference{}, e
	}
	if d.SchemaVersion != 1 || d.ID != ref.SourceID || d.RepositoryID != source.ID || d.Name == "" || d.Quantity == "" || d.Unit == "" || (ref.SourceType == "subobject" && d.ParentObjectID != ref.ParentObjectID) {
		return feature.Reference{}, feature.ErrSourceUnavailable
	}

	raw := object.Content{Name: d.Name, Type: d.Type, Quantity: d.Quantity, Unit: d.Unit, CompletedQuantity: d.CompletedQuantity}
	normalized, normalizeError := object.Normalize(raw)
	if normalizeError != nil || normalized != raw || !object.ValidStatus(d.Status) {
		return feature.Reference{}, feature.ErrSourceUnavailable
	}
	if d.Type == object.TypeWork && d.CompletedQuantity != "" {
		derived, e := object.WorkStatus(d.Quantity, d.CompletedQuantity)
		if e != nil || derived != d.Status {
			return feature.Reference{}, feature.ErrSourceUnavailable
		}
	}
	if ref.SourceType == "subobject" {
		c := child.Content{Name: d.Name, Type: d.Type, Quantity: d.Quantity, Unit: d.Unit, CompletedQuantity: d.CompletedQuantity}
		if _, e := child.Normalize(c); e != nil {
			return feature.Reference{}, feature.ErrSourceUnavailable
		}
		derived, e := child.Status(c)
		if e != nil || derived != d.Status {
			return feature.Reference{}, feature.ErrSourceUnavailable
		}
	}
	ref.SourceVersion = d.Version
	ref.SourceTitle = d.Name
	ref.SourceState = d.Status
	ref.Quantity = d.Quantity
	ref.Unit = d.Unit
	ref.CompletedQuantity = d.CompletedQuantity
	ref.Type = d.Type
	return ref, nil
}

func (s *Store) ListReferences(ctx context.Context, r feature.Scope, actor, pid, token string, hydrate bool) (result feature.ReferencePage, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	root := s.ownerDoc(r.OwnerID)
	parent := root.Collection("projects").Doc(pid)
	scope := []string{"project-references-v1", s.ChannelID, r.OwnerID, actor, pid}
	q, e := applyListCursor(parent.Collection("items").OrderBy(db.DocumentID, db.Asc), token, scope)
	if e != nil {
		return result, e
	}
	authorized := make(map[string]feature.Source)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.ReferencePage{Items: []feature.Reference{}}
		if e := s.checkProjectOwner(tx, r, root); e != nil {
			return e
		}
		snap, e := tx.Get(parent)
		if _, e = decodeProject(snap, e, r); e != nil {
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
			ref := itemReference(d)
			selected := feature.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}
			allowed := hydrate
			if ref.SourceRepositoryID != r.SourceRepositoryID {
				allowed = false
				source, checked := authorized[ref.SourceRepositoryID]
				if !checked && r.Sources != nil {
					var authErr error
					source, authErr = r.Sources.Authorize(ctx, actor, ref.SourceRepositoryID, "")
					if errors.Is(authErr, context.Canceled) || errors.Is(authErr, context.DeadlineExceeded) {
						return authErr
					}
					if authErr != nil {
						source = feature.Source{}
					}
					authorized[ref.SourceRepositoryID] = source
				}
				if source.ID == ref.SourceRepositoryID {
					selected = source
					allowed = true
				}
			} else {
				ref.SourceRepositoryID = r.SourceRepositoryID
			}
			ref.SourceUnavailable = true
			if allowed {
				_, linkErr := tx.Get(parent.Collection("linkedRepositories").Doc(selected.ID))
				if linkErr != nil && status.Code(linkErr) != codes.NotFound {
					return linkErr
				}
				allowed = linkErr == nil
			}
			if allowed {
				current, e := s.sourceReference(tx, selected, ref)
				if e == nil {
					ref = current
					ref.SourceUnavailable = false
				} else if errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) || status.Code(e) == codes.Canceled || status.Code(e) == codes.DeadlineExceeded {
					return e
				}
			}
			result.Items = append(result.Items, ref)
		}
		if more {
			result.NextCursor = firestorekey.EncodeCursor(scope, snaps[len(snaps)-1].Ref.ID)
		}
		return nil
	}, db.ReadOnly)
	if err != nil {
		return feature.ReferencePage{}, err
	}
	return result, nil
}

func (s *Store) ChangeReference(ctx context.Context, r feature.Scope, actor, pid string, ref feature.Reference, add bool, version int64) (result feature.Project, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	root := s.ownerDoc(r.OwnerID)
	parent := root.Collection("projects").Doc(pid)
	itemID := ref.ID
	if add {
		itemID = referenceID(pid, ref)
	}
	item := parent.Collection("items").Doc(itemID)
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
		if actor == "" || actor != current.CreatorID {
			return feature.ErrForbidden
		}
		snap, e = tx.Get(item)
		exists := e == nil
		if e != nil && status.Code(e) != codes.NotFound {
			return e
		}
		if exists {
			d, e := decodeItem(snap, pid)
			if e != nil {
				return e
			}
			if add && referenceID(pid, itemReference(d)) != itemID {
				return feature.ErrUnavailable
			}
		}
		if add {
			selected := feature.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}
			if ref.SourceRepositoryID != r.SourceRepositoryID {
				if r.SelectedSource.ID != ref.SourceRepositoryID {
					return feature.ErrForbidden
				}
				selected = r.SelectedSource
			}
			if _, e = tx.Get(parent.Collection("linkedRepositories").Doc(selected.ID)); e != nil {
				if status.Code(e) == codes.NotFound {
					return feature.ErrForbidden
				}
				return e
			}
			if _, e = s.sourceReference(tx, selected, ref); e != nil {
				if status.Code(e) == codes.NotFound {
					return feature.ErrInput
				}
				return e
			}
		}
		// Identity-idempotent retries remain safe even after the first write bumped version.
		if exists == add {
			return nil
		}
		if e = feature.CheckUpdate(current, actor, version); e != nil {
			return e
		}
		if current.Version >= feature.MaxVersion {
			return feature.ErrConflict
		}
		updated := current
		updated.Version++
		action := "remove_reference"
		if add {
			action = "add_reference"
			d := itemDocument{SourceRepositoryID: ref.SourceRepositoryID, SchemaVersion: 2, ID: itemID, ProjectID: pid, SourceType: ref.SourceType, SourceID: ref.SourceID, ParentObjectID: ref.ParentObjectID, CreatedBy: actor, CreatedAt: now}
			if e = tx.Create(item, d); e != nil {
				return e
			}
		} else {
			if _, exists := snap.Data()["latestInspectionId"]; exists {
				return feature.ErrInspectionHistory
			}
			if e = tx.Delete(item); e != nil {
				return e
			}
		}
		return writeProjectChange(tx, parent, current, updated, actor, action)
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, pid)
}
