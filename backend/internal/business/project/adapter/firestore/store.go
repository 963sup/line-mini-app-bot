package firestore

import (
	projectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/project/application/dto"
	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	orgstore "github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"

	"github.com/963sup/line-mini-app-bot/internal/platform/grpcstatus"

	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"

	db "cloud.google.com/go/firestore"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
)

type (
	Store struct {
		Client    *db.Client
		ChannelID string
	}
	projectDocument struct {
		Settings      projectDomainEntity.Settings `firestore:"settings"`
		SchemaVersion int64                        `firestore:"schemaVersion"`
		ID            string                       `firestore:"id"`

		OwnerID             string    `firestore:"ownerId"`
		DefaultRepositoryID string    `firestore:"defaultRepositoryId"`
		Title               string    `firestore:"title"`
		Body                string    `firestore:"body"`
		CreatorID           string    `firestore:"creatorId"`
		Version             int64     `firestore:"version"`
		CreatedAt           time.Time `firestore:"createdAt"`
		UpdatedAt           time.Time `firestore:"updatedAt"`
	}
)

func encodeProject(i projectDomainEntity.Project) projectDocument {
	return projectDocument{Settings: i.Settings, SchemaVersion: 1, ID: i.ID, OwnerID: i.OwnerID, DefaultRepositoryID: i.DefaultRepositoryID, Title: i.Title, Body: i.Body, CreatorID: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

// Repository-scoped API callers additionally
// restrict results to projects created with that default repository.
func (d projectDocument) belongsTo(scope projectApplicationPort.Scope) bool {
	return (projectDomainEntity.ValidLegacyOwner(d.OwnerID) || projectDomainEntity.IsOrganizationOwner(d.OwnerID)) && d.OwnerID == scope.OwnerID && ((!scope.Legacy && d.DefaultRepositoryID == "") || projectDomainEntity.ValidID(d.DefaultRepositoryID)) &&
		(!scope.Legacy || scope.SourceRepositoryID == "" || scope.SourceRepositoryID == d.DefaultRepositoryID)
}

func decodeProject(snap *db.DocumentSnapshot, err error, scope projectApplicationPort.Scope) (projectDomainEntity.Project, error) {
	if status.Code(err) == codes.NotFound {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrNotFound
	}
	if err != nil {
		return projectDomainEntity.Project{}, err
	}
	var d projectDocument
	if e := snap.DataTo(&d); e != nil {
		return projectDomainEntity.Project{}, e
	}
	c, e := projectDomainEntity.Normalize(projectDomainEntity.Content{Title: d.Title, Body: d.Body})
	if e != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !projectDomainEntity.ValidID(d.ID) || !d.belongsTo(scope) || c.Title != d.Title || c.Body != d.Body || d.CreatorID == "" || d.Version < 1 || d.Version > projectDomainEntity.MaxVersion || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrUnavailable
	}
	return projectDomainEntity.Project{Settings: d.Settings, ID: d.ID, DefaultRepositoryID: d.DefaultRepositoryID, OwnerID: d.OwnerID, Title: d.Title, Body: d.Body, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}

func (s *Store) ownerDoc(id string) *db.DocumentRef {
	collection := "groups"
	if projectDomainEntity.IsOrganizationOwner(id) {
		collection = "organizations"
	}
	if lineid.ValidUser(id) {
		collection = "users"
	}
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection(collection).Doc(id)
}

func (s *Store) repositoryRef(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(id)
}

func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) AuthorizeOrganization(ctx context.Context, actor, owner string) error {
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		return s.checkProjectOwner(tx, projectApplicationPort.Scope{OwnerID: owner, Actor: actor}, s.ownerDoc(owner))
	})
}

func (s *Store) checkProjectOwner(tx *db.Transaction, r projectApplicationPort.Scope, parent *db.DocumentRef) error {
	if lineid.ValidUser(r.OwnerID) {
		if r.OwnerID != r.Actor {
			return projectDomainEntity.ErrForbidden
		}
		return nil
	}
	snap, e := tx.Get(parent)
	if e != nil {
		if status.Code(e) == codes.NotFound {
			return projectDomainEntity.ErrNotFound
		}
		return e
	}
	if projectDomainEntity.IsOrganizationOwner(r.OwnerID) {
		var org orgstore.Document
		if err := snap.DataTo(&org); err != nil {
			return err
		}
		if !org.VisibleLegacyOrganization(r.Actor) {
			return projectDomainEntity.ErrNotFound
		}
		return nil
	}
	var d struct {
		ID, GroupID, Status string
		Version             int64
	}
	if e = snap.DataTo(&d); e != nil {
		return e
	}
	if d.Status != "ACTIVE" || d.Version < 1 {
		return projectDomainEntity.ErrNotFound
	}
	if d.ID != r.OwnerID || d.GroupID != r.OwnerID {
		return projectDomainEntity.ErrNotFound
	}
	return nil
}

func applyListCursor(q db.Query, token string, scope any) (db.Query, error) {
	if token != "" {
		last, ok := firestorekey.DecodeCursor(token, scope)
		if !ok || !projectDomainEntity.ValidID(last) {
			return q, projectDomainEntity.ErrInput
		}
		q = q.StartAfter(last)
	}
	return q.Limit(21), nil
}

func (s *Store) Get(ctx context.Context, r projectApplicationPort.Scope, id string) (result projectDomainEntity.Project, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, projectDomainEntity.ErrUnavailable
	}
	parent := s.ownerDoc(r.OwnerID)
	snap, e := parent.Collection("projects").Doc(id).Get(ctx)
	result, err = decodeProject(snap, e, r)
	if err == nil {
		links, e := snap.Ref.Collection("linkedRepositories").DocumentRefs(ctx).GetAll()
		if e != nil {
			return result, e
		}
		result.LinkedRepositoryIDs = make([]string, 0, len(links))
		for _, link := range links {
			result.LinkedRepositoryIDs = append(result.LinkedRepositoryIDs, link.ID)
		}
	}

	return
}

func (s *Store) List(ctx context.Context, r projectApplicationPort.Scope, actor, token string) (result projectApplicationDto.Page, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	result.Items = []projectDomainEntity.Project{}
	if !s.ready() {
		return result, projectDomainEntity.ErrUnavailable
	}
	parent := s.ownerDoc(r.OwnerID)
	scope := []string{s.ChannelID, r.OwnerID, actor}
	if r.Legacy {
		scope[1] = r.SourceRepositoryID
	}
	q, e := applyListCursor(parent.Collection("projects").OrderBy(db.DocumentID, db.Asc), token, scope)
	if e != nil {
		return result, e
	}
	snaps, e := q.Documents(ctx).GetAll()
	if e != nil {
		return result, e
	}
	more := len(snaps) > 20
	if more {
		snaps = snaps[:20]
	}
	for _, snap := range snaps {
		decodeScope := r
		decodeScope.SourceRepositoryID = ""
		i, e := decodeProject(snap, nil, decodeScope)
		if e != nil {
			return projectApplicationDto.Page{}, e
		}
		if r.Legacy && i.DefaultRepositoryID != r.SourceRepositoryID {
			continue
		}
		result.Items = append(result.Items, i)
	}
	if more {
		result.NextCursor = firestorekey.EncodeCursor(scope, snaps[len(snaps)-1].Ref.ID)
	}
	return result, nil
}

type projectReceipt struct {
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
	SchemaVersion int64  `firestore:"schemaVersion"`
}

func (s *Store) Create(ctx context.Context, r projectApplicationPort.Scope, actor, key string, c projectDomainEntity.Content) (result projectDomainEntity.Project, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, projectDomainEntity.ErrUnavailable
	}
	var random [16]byte
	if _, e := rand.Read(random[:]); e != nil {
		return result, e
	}
	id := hex.EncodeToString(random[:])
	now := time.Now().UTC().Truncate(time.Microsecond)
	payload := firestorekey.JSONDigest(struct {
		Version int    `json:"version"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	}{1, c.Title, c.Body})
	if c.CopyProjectID != "" {
		payload = firestorekey.JSONDigest([]string{payload, c.CopyOwnerID, c.CopyProjectID})
	}
	parent := s.ownerDoc(r.OwnerID)
	creationScope := r.SourceRepositoryID
	if creationScope == "" {
		creationScope = "without-source"
	}
	receipt := parent.Collection("projectCreationScopes").Doc(creationScope).Collection("receipts").Doc(firestorekey.JSONDigest([]string{actor, key}))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.checkProjectOwner(tx, r, parent); e != nil {
			return e
		}
		if r.SourceRepositoryID != "" {
			if e := s.checkSourceRepository(tx, projectDomainEntity.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}); e != nil {
				return e
			}
		}
		var settings projectDomainEntity.Settings
		if c.CopyProjectID != "" {
			sourceOwner := s.ownerDoc(c.CopyOwnerID)
			sourceScope := projectApplicationPort.Scope{OwnerID: c.CopyOwnerID, Actor: actor}
			if err := s.checkProjectOwner(tx, sourceScope, sourceOwner); err != nil {
				return err
			}
			sourceProject := sourceOwner.Collection("projects").Doc(c.CopyProjectID)
			snapshot, err := tx.Get(sourceProject)
			source, err := decodeProject(snapshot, err, sourceScope)
			if err != nil {
				return err
			}
			settings = source.Settings
		}
		snap, e := tx.Get(receipt)
		if e == nil {
			var d projectReceipt
			if e := snap.DataTo(&d); e != nil {
				return e
			}
			if d.SchemaVersion != 1 || !projectDomainEntity.ValidID(d.ID) {
				return projectDomainEntity.ErrUnavailable
			}
			if d.PayloadHash != payload {
				return projectDomainEntity.ErrCreationConflict
			}
			id = d.ID
			return nil
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		i := projectDomainEntity.Project{ID: id, DefaultRepositoryID: r.SourceRepositoryID, OwnerID: r.OwnerID, Title: c.Title, Body: c.Body, CreatorID: actor, Version: 1, CreatedAt: now, UpdatedAt: now}
		i.Settings = settings
		if e := tx.Create(parent.Collection("projects").Doc(id), encodeProject(i)); e != nil {
			return e
		}
		if r.SourceRepositoryID != "" {
			if e := tx.Create(parent.Collection("projects").Doc(id).Collection("linkedRepositories").Doc(r.SourceRepositoryID), map[string]string{"repositoryId": r.SourceRepositoryID}); e != nil {
				return e
			}
		}
		return tx.Create(receipt, projectReceipt{id, payload, 1})
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, id)
}

type projectChange struct {
	ActorID string    `firestore:"actorId"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
	Action  string    `firestore:"action"`
}

func writeProjectChange(tx *db.Transaction, ref *db.DocumentRef, current, result projectDomainEntity.Project, actor, action string) error {
	if result.Version == current.Version {
		return nil
	}
	result.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if result.UpdatedAt.Before(current.UpdatedAt) {
		result.UpdatedAt = current.UpdatedAt
	}
	if e := tx.Set(ref, encodeProject(result)); e != nil {
		return e
	}
	return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), projectChange{actor, result.Version, result.UpdatedAt, action})
}

func (s *Store) RunUpdate(ctx context.Context, r projectApplicationPort.Scope, id string, update func(projectApplicationPort.UpdateWork) error) (result projectDomainEntity.Project, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, projectDomainEntity.ErrUnavailable
	}
	if r.SourceRepositoryID == "" {
		value, e := s.Get(ctx, r, id)
		if e != nil {
			return result, e
		}
		r.SourceRepositoryID = value.DefaultRepositoryID
	}
	parent := s.ownerDoc(r.OwnerID)
	ref := parent.Collection("projects").Doc(id)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.checkProjectOwner(tx, r, parent); e != nil {
			return e
		}
		return update(projectUpdateWork{store: s, tx: tx, ref: ref, scope: r})
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, id)
}

type projectUpdateWork struct {
	store *Store
	tx    *db.Transaction
	ref   *db.DocumentRef
	scope projectApplicationPort.Scope
}

func (w projectUpdateWork) Current() (projectDomainEntity.Project, error) {
	snap, err := w.tx.Get(w.ref)
	return decodeProject(snap, err, w.scope)
}

func (w projectUpdateWork) ItemStatus(id string) (string, error) {
	snap, err := w.tx.Get(w.ref.Collection("items").Doc(id))
	if status.Code(err) == codes.NotFound {
		return "", projectDomainEntity.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	doc, err := decodeItem(snap, w.ref.ID)
	if err != nil {
		return "", err
	}
	return itemReference(doc).ItemStatus, nil
}

func (w projectUpdateWork) RepositoryLinked(id string) (bool, error) {
	_, err := w.tx.Get(w.ref.Collection("linkedRepositories").Doc(id))
	if status.Code(err) == codes.NotFound {
		return false, nil
	}
	return err == nil, err
}

func (w projectUpdateWork) CheckLinkSource(id string) error {
	repo, err := w.tx.Get(w.store.repositoryRef(id))
	if err != nil {
		return err
	}
	var src struct{ ID, OwnerID, Status string }
	if err = repo.DataTo(&src); err != nil {
		return err
	}
	if src.ID != id || src.OwnerID != w.scope.SelectedSource.OwnerID || src.Status != "ACTIVE" {
		return projectDomainEntity.ErrForbidden
	}
	return nil
}

func (w projectUpdateWork) Save(actor string, previous, updated projectDomainEntity.Project, changes projectApplicationPort.UpdateWrites) error {
	if changes.ItemID != "" {
		item := w.ref.Collection("items").Doc(changes.ItemID)
		if err := w.tx.Update(item, []db.Update{{Path: "itemStatus", Value: changes.ItemStatus}}); err != nil {
			return err
		}
	}
	if changes.RepositoryID != "" {
		link := w.ref.Collection("linkedRepositories").Doc(changes.RepositoryID)
		if changes.LinkRepository {
			if err := w.tx.Create(link, map[string]string{"repositoryId": changes.RepositoryID}); err != nil {
				return err
			}
		} else if err := w.tx.Delete(link); err != nil {
			return err
		}
	}
	return writeProjectChange(w.tx, w.ref, previous, updated, actor, "update")
}
