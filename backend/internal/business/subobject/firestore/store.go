package firestore

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/subobject"
	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

var _ feature.Store = (*Store)(nil)

type document struct {
	SchemaVersion     int64     `firestore:"schemaVersion"`
	ID                string    `firestore:"id"`
	RepositoryID      string    `firestore:"repositoryId"`
	ParentObjectID    string    `firestore:"parentObjectId"`
	CreatorID         string    `firestore:"creatorId"`
	Name              string    `firestore:"name"`
	Type              string    `firestore:"type"`
	Quantity          string    `firestore:"quantity"`
	Unit              string    `firestore:"unit"`
	CompletedQuantity string    `firestore:"completedQuantity,omitempty"`
	Status            string    `firestore:"status"`
	Version           int64     `firestore:"version"`
	CreatedAt         time.Time `firestore:"createdAt"`
	UpdatedAt         time.Time `firestore:"updatedAt"`
}

func encode(j feature.Subobject) document {
	return document{SchemaVersion: 1, ID: j.ID, RepositoryID: j.RepositoryID, ParentObjectID: j.ParentObjectID, CreatorID: j.CreatorID, Name: j.Name, Type: j.Type, Quantity: j.Quantity, Unit: j.Unit, CompletedQuantity: j.CompletedQuantity, Status: j.Status, Version: j.Version, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt}
}

func decode(snap *db.DocumentSnapshot, e error, repositoryID, parentObjectID, parentCreatorID string) (feature.Subobject, error) {
	if status.Code(e) == codes.NotFound {
		return feature.Subobject{}, feature.ErrNotFound
	}
	if e != nil {
		return feature.Subobject{}, e
	}
	var d document
	if e = snap.DataTo(&d); e != nil {
		return feature.Subobject{}, e
	}
	c := feature.Content{Name: d.Name, Type: d.Type, Quantity: d.Quantity, Unit: d.Unit, CompletedQuantity: d.CompletedQuantity}
	normalized, e := feature.Normalize(c)
	derived, se := feature.Status(c)
	if e != nil || se != nil || normalized != c || derived != d.Status || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != repositoryID || d.ParentObjectID != parentObjectID || d.CreatorID != parentCreatorID || parentCreatorID == "" || d.Version < 1 || d.Version > repository.MaxResourceNumber || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return feature.Subobject{}, feature.ErrUnavailable
	}
	return feature.Subobject{Content: c, ID: d.ID, RepositoryID: repositoryID, ParentObjectID: parentObjectID, CreatorID: d.CreatorID, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) repositoryRef(repositoryID string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(repositoryID)
}

func (s *Store) parentObjectRef(repositoryID, parentObjectID string) *db.DocumentRef {
	return s.repositoryRef(repositoryID).Collection("objects").Doc(parentObjectID)
}

// authorizeParentObject returns the parent creator ID and restricts writes to that creator.
func (s *Store) authorizeParentObject(tx *db.Transaction, r repository.Repository, parentObjectID, actor string, requireCreator bool) (string, error) {
	snap, e := tx.Get(s.repositoryRef(r.ID))
	if status.Code(e) == codes.NotFound {
		return "", feature.ErrNotFound
	}
	if e != nil {
		return "", e
	}
	var repo struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	if e = snap.DataTo(&repo); e != nil {
		return "", e
	}
	if !r.MatchesActive(repo.ID, repo.OwnerID, repo.Status) {
		return "", feature.ErrNotFound
	}
	snap, e = tx.Get(s.parentObjectRef(r.ID, parentObjectID))
	if status.Code(e) == codes.NotFound {
		return "", feature.ErrNotFound
	}
	if e != nil {
		return "", e
	}
	var parent struct {
		ID            string `firestore:"id"`
		RepositoryID  string `firestore:"repositoryId"`
		CreatorID     string `firestore:"creatorId"`
		SchemaVersion int64  `firestore:"schemaVersion"`
	}
	if e = snap.DataTo(&parent); e != nil {
		return "", e
	}
	if parent.ID != parentObjectID || parent.RepositoryID != r.ID || parent.CreatorID == "" || parent.SchemaVersion != 1 {
		return "", feature.ErrNotFound
	}
	if requireCreator && parent.CreatorID != actor {
		return "", feature.ErrForbidden
	}
	return parent.CreatorID, nil
}

func (s *Store) Get(ctx context.Context, r repository.Repository, actor, parentObjectID, subobjectID string) (result feature.Subobject, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Subobject{}
		parentCreatorID, e := s.authorizeParentObject(tx, r, parentObjectID, actor, false)
		if e != nil {
			return e
		}
		snap, e := tx.Get(s.parentObjectRef(r.ID, parentObjectID).Collection("subobjects").Doc(subobjectID))
		result, e = decode(snap, e, r.ID, parentObjectID, parentCreatorID)
		return e
	})
	if err != nil {
		return feature.Subobject{}, err
	}
	return result, nil
}

type cursor struct{ Scope, ID string }

func (s *Store) List(ctx context.Context, r repository.Repository, actor, parentObjectID, token string) (result feature.Page, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	scope := firestorekey.JSONDigest([]string{"subobject-cursor-v1", s.ChannelID, r.ID, parentObjectID, actor})
	c := cursor{}
	if token != "" {
		b, e := base64.RawURLEncoding.DecodeString(token)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repository.ValidID(c.ID) {
			return result, feature.ErrInput
		}
	}
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Page{Items: []feature.Subobject{}}
		parentCreatorID, e := s.authorizeParentObject(tx, r, parentObjectID, actor, false)
		if e != nil {
			return e
		}
		q := s.parentObjectRef(r.ID, parentObjectID).Collection("subobjects").OrderBy(db.DocumentID, db.Asc).Limit(21)
		if c.ID != "" {
			q = q.StartAfter(c.ID)
		}
		docs, e := tx.Documents(q).GetAll()
		if e != nil {
			return e
		}
		more := len(docs) > 20
		if more {
			docs = docs[:20]
		}
		for _, snap := range docs {
			j, e := decode(snap, nil, r.ID, parentObjectID, parentCreatorID)
			if e != nil {
				return e
			}
			result.Items = append(result.Items, j)
		}
		if more {
			b, _ := json.Marshal(cursor{Scope: scope, ID: docs[len(docs)-1].Ref.ID})
			result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
		}
		return nil
	})
	if err != nil {
		return feature.Page{}, err
	}
	return result, nil
}

type receipt struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
}

func (s *Store) Create(ctx context.Context, r repository.Repository, actor, parentObjectID, key string, c feature.Content) (result feature.Subobject, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	parent := s.parentObjectRef(r.ID, parentObjectID)
	ref := parent.Collection("subobjects").NewDoc() // IDs follow the same opaque 32-hex convention as Object.
	subobjectID := firestorekey.JSONDigest([]string{"subobject-id-v1", ref.ID})[:32]
	ref = parent.Collection("subobjects").Doc(subobjectID)
	rec := parent.Collection("subobjectCreationReceipts").Doc(firestorekey.JSONDigest([]string{"subobject-create-v1", actor, parentObjectID, key}))
	payload := firestorekey.JSONDigest(struct {
		Version                                       int
		Name, Type, Quantity, Unit, CompletedQuantity string
	}{1, c.Name, c.Type, c.Quantity, c.Unit, c.CompletedQuantity})
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Subobject{}
		parentCreatorID, e := s.authorizeParentObject(tx, r, parentObjectID, actor, true)
		if e != nil {
			return e
		}
		snap, e := tx.Get(rec)
		if e == nil {
			var old receipt
			if e = snap.DataTo(&old); e != nil {
				return e
			}
			if old.SchemaVersion != 1 || !repository.ValidID(old.ID) || len(old.PayloadHash) != 64 {
				return feature.ErrUnavailable
			}
			if old.PayloadHash != payload {
				return feature.ErrCreationConflict
			}
			snap, e = tx.Get(parent.Collection("subobjects").Doc(old.ID))
			result, e = decode(snap, e, r.ID, parentObjectID, parentCreatorID)
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		derived, e := feature.Status(c)
		if e != nil {
			return e
		}
		result = feature.Subobject{Content: c, ID: subobjectID, RepositoryID: r.ID, ParentObjectID: parentObjectID, CreatorID: parentCreatorID, Status: derived, Version: 1, CreatedAt: now, UpdatedAt: now}
		if e = tx.Create(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(rec, receipt{1, subobjectID, payload})
	})
	if err != nil {
		return feature.Subobject{}, err
	}
	return result, nil
}

func (s *Store) Update(ctx context.Context, r repository.Repository, actor, parentObjectID, subobjectID string, input feature.Update) (result feature.Subobject, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	ref := s.parentObjectRef(r.ID, parentObjectID).Collection("subobjects").Doc(subobjectID)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = feature.Subobject{}
		parentCreatorID, e := s.authorizeParentObject(tx, r, parentObjectID, actor, true)
		if e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		old, e := decode(snap, e, r.ID, parentObjectID, parentCreatorID)
		if e != nil {
			return e
		}
		result, e = feature.ApplyUpdate(old, actor, input)
		if e != nil {
			return e
		}
		if result.Version == old.Version {
			return nil
		}
		result.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		if !result.UpdatedAt.After(old.UpdatedAt) {
			result.UpdatedAt = old.UpdatedAt.Add(time.Microsecond)
		}
		if e = tx.Set(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), map[string]any{"actorId": actor, "version": result.Version, "at": result.UpdatedAt, "before": encode(old), "after": encode(result)})
	})
	if err != nil {
		return feature.Subobject{}, err
	}
	return result, nil
}
