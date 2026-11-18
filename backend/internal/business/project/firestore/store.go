package firestore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"

	db "cloud.google.com/go/firestore"
	feature "github.com/963sup/line-mini-app-bot/internal/business/project"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	Store struct {
		Client    *db.Client
		ChannelID string
	}
	planningDocument struct {
		SchemaVersion int64  `firestore:"schemaVersion"`
		ID            string `firestore:"id"`

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

func planningEncode(i feature.Project) planningDocument {
	return planningDocument{SchemaVersion: 1, ID: i.ID, OwnerID: i.OwnerID, DefaultRepositoryID: i.DefaultRepositoryID, Title: i.Title, Body: i.Body, CreatorID: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}

// Repository-scoped API callers additionally
// restrict results to projects created with that default repository.
func (d planningDocument) belongsTo(scope feature.Scope) bool {
	return identity.ValidOwner(d.OwnerID) && d.OwnerID == scope.OwnerID && feature.ValidID(d.DefaultRepositoryID) &&
		(!scope.Legacy || scope.SourceRepositoryID == "" || scope.SourceRepositoryID == d.DefaultRepositoryID)
}

func planningDecode(snap *db.DocumentSnapshot, err error, scope feature.Scope) (feature.Project, error) {
	if status.Code(err) == codes.NotFound {
		return feature.Project{}, feature.ErrNotFound
	}
	if err != nil {
		return feature.Project{}, err
	}
	var d planningDocument
	if e := snap.DataTo(&d); e != nil {
		return feature.Project{}, e
	}
	c, e := feature.Normalize(feature.Content{Title: d.Title, Body: d.Body})
	if e != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !feature.ValidID(d.ID) || !d.belongsTo(scope) || c.Title != d.Title || c.Body != d.Body || d.CreatorID == "" || d.Version < 1 || d.Version > feature.MaxVersion || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return feature.Project{}, feature.ErrUnavailable
	}
	return feature.Project{ID: d.ID, DefaultRepositoryID: d.DefaultRepositoryID, OwnerID: d.OwnerID, Title: d.Title, Body: d.Body, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}

type planningFailure struct{ cause error }

func (e *planningFailure) Error() string { return "project storage unavailable" }
func (e *planningFailure) Unwrap() error { return feature.ErrUnavailable }
func planningError(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{feature.ErrInput, feature.ErrNotFound, feature.ErrForbidden, feature.ErrConflict, feature.ErrCreationConflict, feature.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, known) {
			return err
		}
	}
	if status.Code(err) == codes.Canceled {
		return context.Canceled
	}
	if status.Code(err) == codes.DeadlineExceeded {
		return context.DeadlineExceeded
	}
	return &planningFailure{err}
}

func (s *Store) ownerDoc(id string) *db.DocumentRef {
	collection := "groups"
	if identity.IsUserOwnerID(id) {
		collection = "users"
	}
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection(collection).Doc(id)
}

func (s *Store) repositoryRef(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(id)
}

func (s *Store) planningReady() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) planningContainer(tx *db.Transaction, r feature.Scope, parent *db.DocumentRef) error {
	if identity.IsUserOwnerID(r.OwnerID) {
		if r.OwnerID != r.Actor {
			return feature.ErrForbidden
		}
		return nil
	}
	snap, e := tx.Get(parent)
	if e != nil {
		if status.Code(e) == codes.NotFound {
			return feature.ErrNotFound
		}
		return e
	}
	var d struct {
		ID, GroupID, Status string
		Version             int64
	}
	if e = snap.DataTo(&d); e != nil {
		return e
	}
	if d.Status != "ACTIVE" || d.Version < 1 {
		return feature.ErrNotFound
	}
	if d.ID != r.OwnerID || d.GroupID != r.OwnerID {
		return feature.ErrNotFound
	}
	return nil
}

func planningHash(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

type planningCursor struct{ Scope, Last string }

func planningPage(q db.Query, token string, scope any) (db.Query, error) {
	if token != "" {
		b, e := base64.RawURLEncoding.DecodeString(token)
		var c planningCursor
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != planningHash(scope) || !feature.ValidID(c.Last) {
			return q, feature.ErrInput
		}
		q = q.StartAfter(c.Last)
	}
	return q.Limit(21), nil
}

func planningNext(scope any, id string) string {
	b, _ := json.Marshal(planningCursor{planningHash(scope), id})
	return base64.RawURLEncoding.EncodeToString(b)
}

func (s *Store) Get(ctx context.Context, r feature.Scope, id string) (result feature.Project, err error) {
	defer func() { err = planningError(err) }()
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	parent := s.ownerDoc(r.OwnerID)
	snap, e := parent.Collection("projects").Doc(id).Get(ctx)
	result, err = planningDecode(snap, e, r)
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

func (s *Store) List(ctx context.Context, r feature.Scope, actor, token string) (result feature.Page, err error) {
	defer func() { err = planningError(err) }()
	result.Items = []feature.Project{}
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	parent := s.ownerDoc(r.OwnerID)
	scope := []string{s.ChannelID, r.OwnerID, actor}
	if r.Legacy {
		scope[1] = r.SourceRepositoryID
	}
	q, e := planningPage(parent.Collection("projects").OrderBy(db.DocumentID, db.Asc), token, scope)
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
		i, e := planningDecode(snap, nil, decodeScope)
		if e != nil {
			return feature.Page{}, e
		}
		if r.Legacy && i.DefaultRepositoryID != r.SourceRepositoryID {
			continue
		}
		result.Items = append(result.Items, i)
	}
	if more {
		result.NextCursor = planningNext(scope, snaps[len(snaps)-1].Ref.ID)
	}
	return result, nil
}

type planningReceipt struct {
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
	SchemaVersion int64  `firestore:"schemaVersion"`
}

func (s *Store) Create(ctx context.Context, r feature.Scope, actor, key string, c feature.Content) (result feature.Project, err error) {
	defer func() { err = planningError(err) }()
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	var random [16]byte
	if _, e := rand.Read(random[:]); e != nil {
		return result, e
	}
	id := hex.EncodeToString(random[:])
	now := time.Now().UTC().Truncate(time.Microsecond)
	payload := planningHash(struct {
		Version int    `json:"version"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	}{1, c.Title, c.Body})
	parent := s.ownerDoc(r.OwnerID)
	receipt := parent.Collection("projectCreationScopes").Doc(r.SourceRepositoryID).Collection("receipts").Doc(planningHash([]string{actor, key}))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.planningContainer(tx, r, parent); e != nil {
			return e
		}
		if e := s.checkSourceRepository(tx, feature.Source{ID: r.SourceRepositoryID, OwnerID: r.OwnerID}); e != nil {
			return e
		}
		snap, e := tx.Get(receipt)
		if e == nil {
			var d planningReceipt
			if e := snap.DataTo(&d); e != nil {
				return e
			}
			if d.SchemaVersion != 1 || !feature.ValidID(d.ID) {
				return feature.ErrUnavailable
			}
			if d.PayloadHash != payload {
				return feature.ErrCreationConflict
			}
			id = d.ID
			return nil
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		i := feature.Project{ID: id, DefaultRepositoryID: r.SourceRepositoryID, OwnerID: r.OwnerID, Title: c.Title, Body: c.Body, CreatorID: actor, Version: 1, CreatedAt: now, UpdatedAt: now}
		if e := tx.Create(parent.Collection("projects").Doc(id), planningEncode(i)); e != nil {
			return e
		}
		if e := tx.Create(parent.Collection("projects").Doc(id).Collection("linkedRepositories").Doc(r.SourceRepositoryID), map[string]string{"repositoryId": r.SourceRepositoryID}); e != nil {
			return e
		}
		return tx.Create(receipt, planningReceipt{id, payload, 1})
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, id)
}

type planningChange struct {
	ActorID string    `firestore:"actorId"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
	Action  string    `firestore:"action"`
}

func (s *Store) planningWrite(tx *db.Transaction, ref *db.DocumentRef, current, result feature.Project, actor, action string) error {
	if result.Version == current.Version {
		return nil
	}
	result.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if result.UpdatedAt.Before(current.UpdatedAt) {
		result.UpdatedAt = current.UpdatedAt
	}
	if e := tx.Set(ref, planningEncode(result)); e != nil {
		return e
	}
	return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), planningChange{actor, result.Version, result.UpdatedAt, action})
}

func (s *Store) Update(ctx context.Context, r feature.Scope, actor, id string, input feature.Update) (result feature.Project, err error) {
	defer func() { err = planningError(err) }()
	if !s.planningReady() {
		return result, feature.ErrUnavailable
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
		if e := s.planningContainer(tx, r, parent); e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		current, e := planningDecode(snap, e, r)
		if e != nil {
			return e
		}
		updated, e := feature.ApplyUpdate(current, actor, input)
		if e != nil {
			return e
		}
		linkID, add := input.LinkRepositoryID, true
		if linkID == "" {
			linkID, add = input.UnlinkRepositoryID, false
		}
		if linkID != "" {
			link := ref.Collection("linkedRepositories").Doc(linkID)
			_, err := tx.Get(link)
			exists := err == nil
			if err != nil && status.Code(err) != codes.NotFound {
				return err
			}
			if exists != add {
				if current.Version >= feature.MaxVersion {
					return feature.ErrConflict
				}
				if add {
					repo, err := tx.Get(s.repositoryRef(linkID))
					if err != nil {
						return err
					}
					var src struct{ ID, OwnerID, Status string }
					if err = repo.DataTo(&src); err != nil {
						return err
					}
					if src.ID != linkID || src.OwnerID != r.SelectedSource.OwnerID || src.Status != "ACTIVE" {
						return feature.ErrForbidden
					}
					if err := tx.Create(link, map[string]string{"repositoryId": linkID}); err != nil {
						return err
					}
				} else if err := tx.Delete(link); err != nil {
					return err
				}
				updated.Version = current.Version + 1
			}
		}
		return s.planningWrite(tx, ref, current, updated, actor, "update")
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, id)
}
