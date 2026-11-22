package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	objectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/object/application/dto"
	objectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/object/application/port"
	objectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/object/domain/entity"
)

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	db "cloud.google.com/go/firestore"

	"github.com/963sup/line-mini-app-bot/internal/platform/grpcstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

var _ objectApplicationPort.Store = (*Store)(nil)

type document struct {
	TemplateID        string    `firestore:"templateId,omitempty"`
	TemplateNodeID    string    `firestore:"templateNodeId,omitempty"`
	TemplateVersion   int64     `firestore:"templateVersion,omitempty"`
	ParentID          string    `firestore:"parentId,omitempty"`
	SpaceID           string    `firestore:"spaceId,omitempty"`
	CompletedQuantity string    `firestore:"completedQuantity,omitempty"`
	Name              string    `firestore:"name"`
	Description       string    `firestore:"description"`
	Type              string    `firestore:"type"`
	Quantity          string    `firestore:"quantity"`
	Unit              string    `firestore:"unit"`
	SourceDocument    string    `firestore:"sourceDocument"`
	SourceRevision    string    `firestore:"sourceRevision"`
	SourceItemNo      string    `firestore:"sourceItemNo"`
	SourcePages       string    `firestore:"sourcePages"`
	MaterialCode      string    `firestore:"materialCode"`
	SchemaVersion     int64     `firestore:"schemaVersion"`
	ID                string    `firestore:"id"`
	RepositoryID      string    `firestore:"repositoryId"`
	CreatorID         string    `firestore:"creatorId"`
	Status            string    `firestore:"status"`
	Version           int64     `firestore:"version"`
	CreatedAt         time.Time `firestore:"createdAt"`
	UpdatedAt         time.Time `firestore:"updatedAt"`
}

func encode(i objectDomainEntity.Object) document {
	return document{TemplateID: i.TemplateID, TemplateNodeID: i.TemplateNodeID, TemplateVersion: i.TemplateVersion, ParentID: i.ParentID, SpaceID: i.SpaceID, CompletedQuantity: i.CompletedQuantity, Name: i.Name, Description: i.Description, Type: i.Type, Quantity: i.Quantity, Unit: i.Unit, SourceDocument: i.SourceDocument, SourceRevision: i.SourceRevision, SourceItemNo: i.SourceItemNo, SourcePages: i.SourcePages, MaterialCode: i.MaterialCode, SchemaVersion: 1, ID: i.ID, RepositoryID: i.RepositoryID, CreatorID: i.CreatorID, Status: i.Status, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

func decode(snap *db.DocumentSnapshot, err error, rid string) (objectDomainEntity.Object, error) {
	if status.Code(err) == codes.NotFound {
		return objectDomainEntity.Object{}, objectDomainEntity.ErrNotFound
	}
	if err != nil {
		return objectDomainEntity.Object{}, err
	}
	var d document
	if e := snap.DataTo(&d); e != nil {
		return objectDomainEntity.Object{}, e
	}
	raw := objectDomainEntity.Content{ParentID: d.ParentID, SpaceID: d.SpaceID, CompletedQuantity: d.CompletedQuantity, Name: d.Name, Description: d.Description, Type: d.Type, Quantity: d.Quantity, Unit: d.Unit, SourceDocument: d.SourceDocument, SourceRevision: d.SourceRevision, SourceItemNo: d.SourceItemNo, SourcePages: d.SourcePages, MaterialCode: d.MaterialCode}
	c, e := objectDomainEntity.Normalize(raw)
	if e != nil || c != raw || d.SchemaVersion != 1 || d.ID == d.ParentID || d.ID != snap.Ref.ID || !repositoryDomainEntity.ValidID(d.ID) || d.RepositoryID != rid || d.CreatorID == "" || !objectDomainEntity.ValidStatus(d.Status) || d.Version < 1 || d.Version > repositoryDomainEntity.MaxResourceNumber || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return objectDomainEntity.Object{}, objectDomainEntity.ErrUnavailable
	}
	if c.Type == objectDomainEntity.TypeWork && c.CompletedQuantity != "" {
		derived, e := objectDomainEntity.WorkStatus(c.Quantity, c.CompletedQuantity)
		if e != nil || derived != d.Status {
			return objectDomainEntity.Object{}, objectDomainEntity.ErrUnavailable
		}
	}
	return objectDomainEntity.Object{TemplateID: d.TemplateID, TemplateNodeID: d.TemplateNodeID, TemplateVersion: d.TemplateVersion, Content: c, ID: d.ID, RepositoryID: d.RepositoryID, CreatorID: d.CreatorID, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) repositoryRef(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(id)
}

func translate(e error) error {
	if e == nil {
		return nil
	}
	for _, v := range []error{objectDomainEntity.ErrProgressRemoval, objectDomainEntity.ErrInput, objectDomainEntity.ErrNotFound, objectDomainEntity.ErrForbidden, objectDomainEntity.ErrConflict, objectDomainEntity.ErrCreationConflict, objectDomainEntity.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(e, v) {
			return e
		}
	}
	return grpcstatus.NormalizeContextError(e)
}

func (s *Store) Get(ctx context.Context, r repositoryDomainEntity.Repository, id string) (result objectDomainEntity.Object, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, objectDomainEntity.ErrUnavailable
	}
	snap, e := s.repositoryRef(r.ID).Collection("objects").Doc(id).Get(ctx)
	return decode(snap, e, r.ID)
}

type cursor struct{ Scope, ID string }

func (s *Store) List(ctx context.Context, r repositoryDomainEntity.Repository, actor, token string, parent ...string) (result objectApplicationDto.Page, err error) {
	defer func() { err = translate(err) }()
	result.Items = []objectDomainEntity.Object{}
	if !s.ready() {
		return result, objectDomainEntity.ErrUnavailable
	}
	scopeParts := []string{s.ChannelID, r.ID, actor}
	if len(parent) > 0 && parent[0] != "" {
		scopeParts = append(scopeParts, parent[0])
	}
	scope := firestorekey.JSONDigest(scopeParts)
	c := cursor{}
	if token != "" {
		b, e := base64.RawURLEncoding.DecodeString(token)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repositoryDomainEntity.ValidID(c.ID) {
			return result, objectDomainEntity.ErrInput
		}
	}
	q := s.repositoryRef(r.ID).Collection("objects").OrderBy(db.DocumentID, db.Asc)
	if len(parent) > 0 && parent[0] != "" {
		q = q.Where("parentId", "==", parent[0])
	}
	if c.ID != "" {
		q = q.StartAfter(c.ID)
	}
	docs, e := q.Limit(21).Documents(ctx).GetAll()
	if e != nil {
		return result, e
	}
	more := len(docs) > 20
	if more {
		docs = docs[:20]
	}
	for _, snap := range docs {
		i, e := decode(snap, nil, r.ID)
		if e != nil {
			return objectApplicationDto.Page{}, e
		}
		result.Items = append(result.Items, i)
		c = cursor{scope, i.ID}
	}
	if more {
		b, _ := json.Marshal(c)
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func container(tx *db.Transaction, parent *db.DocumentRef, r repositoryDomainEntity.Repository) error {
	snap, e := tx.Get(parent)
	if status.Code(e) == codes.NotFound {
		return objectDomainEntity.ErrNotFound
	}
	if e != nil {
		return e
	}
	var d struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	if e = snap.DataTo(&d); e != nil {
		return e
	}
	if !r.MatchesActive(d.ID, d.OwnerID, d.Status) {
		return objectDomainEntity.ErrNotFound
	}
	return nil
}

type receipt struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
}

func (s *Store) Create(ctx context.Context, r repositoryDomainEntity.Repository, actor, key string, c objectDomainEntity.Content) (result objectDomainEntity.Object, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, objectDomainEntity.ErrUnavailable
	}
	var random [16]byte
	if _, e := rand.Read(random[:]); e != nil {
		return result, e
	}
	id := hex.EncodeToString(random[:])
	now := time.Now().UTC().Truncate(time.Microsecond)
	parent := s.repositoryRef(r.ID)
	ref := parent.Collection("objects").Doc(id)
	// Unnumbered Object receipts use their own namespace; no counter or number allocation.
	rec := parent.Collection("objectCreationReceipts").Doc(firestorekey.JSONDigest([]string{actor, key}))
	payload := creationHash(c)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = objectDomainEntity.Object{}
		if e := container(tx, parent, r); e != nil {
			return e
		}
		if e := validateSpace(tx, parent, r.ID, c.SpaceID); e != nil {
			return e
		}
		prior, e := tx.Get(rec)
		if e == nil {
			var old receipt
			if e = prior.DataTo(&old); e != nil {
				return e
			}
			if old.SchemaVersion != 1 || !repositoryDomainEntity.ValidID(old.ID) || len(old.PayloadHash) != 64 {
				return objectDomainEntity.ErrUnavailable
			}
			if old.PayloadHash != payload {
				return objectDomainEntity.ErrCreationConflict
			}
			snap, e := tx.Get(parent.Collection("objects").Doc(old.ID))
			result, e = decode(snap, e, r.ID)
			if e == nil && result.CreatorID != actor {
				return objectDomainEntity.ErrUnavailable
			}
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		// Parent links are immutable and only point to existing objects in this repository.
		if c.ParentID != "" {
			snap, parentErr := tx.Get(parent.Collection("objects").Doc(c.ParentID))
			if _, parentErr = decode(snap, parentErr, r.ID); parentErr != nil {
				return parentErr
			}
		}
		result = objectDomainEntity.Object{Content: c, ID: id, RepositoryID: r.ID, CreatorID: actor, Status: initialStatus(c), Version: 1, CreatedAt: now, UpdatedAt: now}
		if e = tx.Create(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(rec, receipt{1, id, payload})
	})
	if err != nil {
		return objectDomainEntity.Object{}, err
	}
	return result, nil
}

func (s *Store) Update(ctx context.Context, r repositoryDomainEntity.Repository, actor, id string, u objectDomainEntity.Update) (result objectDomainEntity.Object, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, objectDomainEntity.ErrUnavailable
	}
	parent := s.repositoryRef(r.ID)
	ref := parent.Collection("objects").Doc(id)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = objectDomainEntity.Object{}
		if e := container(tx, parent, r); e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		current, e := decode(snap, e, r.ID)
		if e != nil {
			return e
		}
		result, e = objectDomainEntity.ApplyUpdate(current, actor, u)
		if e != nil {
			return e
		}
		if e := validateSpace(tx, parent, r.ID, result.SpaceID); e != nil {
			return e
		}
		if result.Version == current.Version {
			return nil
		}
		result.UpdatedAt = now
		if !result.UpdatedAt.After(current.UpdatedAt) {
			result.UpdatedAt = current.UpdatedAt.Add(time.Microsecond)
		}
		if e = tx.Set(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), map[string]any{"actorId": actor, "version": result.Version, "at": result.UpdatedAt, "before": encode(current), "after": encode(result)})
	})
	if err != nil {
		return objectDomainEntity.Object{}, err
	}
	return result, nil
}

func initialStatus(c objectDomainEntity.Content) string {
	if c.Type == objectDomainEntity.TypeWork && c.CompletedQuantity != "" {
		status, _ := objectDomainEntity.WorkStatus(c.Quantity, c.CompletedQuantity)
		return status
	}
	return objectDomainEntity.StatusNotStarted
}

// validateSpace reads only below the current channel/repository parent. Business
// Space membership never grants permission independently of the Repository.
func validateSpace(tx *db.Transaction, parent *db.DocumentRef, rid, id string) error {
	if id == "" {
		return nil
	}
	if !repositoryDomainEntity.ValidID(id) {
		return objectDomainEntity.ErrInput
	}
	snap, err := tx.Get(parent.Collection("spaces").Doc(id))
	if status.Code(err) == codes.NotFound {
		return objectDomainEntity.ErrInput
	}
	if err != nil {
		return err
	}
	var d struct {
		SchemaVersion int64  `firestore:"schemaVersion"`
		ID            string `firestore:"id"`
		RepositoryID  string `firestore:"repositoryId"`
	}
	if err := snap.DataTo(&d); err != nil {
		return err
	}
	if d.SchemaVersion != 1 || d.ID != id || d.RepositoryID != rid {
		return objectDomainEntity.ErrInput
	}
	return nil
}

// Preserve immutable v1 receipts for every legacy input; progress uses a new intent shape.
func creationHash(c objectDomainEntity.Content) string {
	if c.ParentID != "" {
		parentID := c.ParentID
		c.ParentID = ""
		return firestorekey.JSONDigest(struct {
			Version               int
			ContentHash, ParentID string
		}{4, creationHash(c), parentID})
	}
	if c.SpaceID != "" {
		spaceID := c.SpaceID
		c.SpaceID = ""
		return firestorekey.JSONDigest(struct {
			Version              int
			ContentHash, SpaceID string
		}{3, creationHash(c), spaceID})
	}
	type legacyContent struct{ Name, Description, Type, Quantity, Unit, SourceDocument, SourceRevision, SourceItemNo, SourcePages, MaterialCode string }
	old := legacyContent{c.Name, c.Description, c.Type, c.Quantity, c.Unit, c.SourceDocument, c.SourceRevision, c.SourceItemNo, c.SourcePages, c.MaterialCode}
	if c.CompletedQuantity == "" || (c.Type == objectDomainEntity.TypeWork && c.CompletedQuantity == "0") {
		return firestorekey.JSONDigest(struct {
			Version int
			Content legacyContent
		}{1, old})
	}
	return firestorekey.JSONDigest(struct {
		Version           int
		Content           legacyContent
		CompletedQuantity string
	}{2, old, c.CompletedQuantity})
}
