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
		SchemaVersion      int64     `firestore:"schemaVersion"`
		ID                 string    `firestore:"id"`
		RepositoryID       string    `firestore:"repositoryId"`
		SpaceID            string    `firestore:"spaceId,omitempty"`
		SourceRepositoryID string    `firestore:"sourceRepositoryId,omitempty"`
		Title              string    `firestore:"title"`
		Body               string    `firestore:"body"`
		CreatorID          string    `firestore:"creatorId"`
		Version            int64     `firestore:"version"`
		CreatedAt          time.Time `firestore:"createdAt"`
		UpdatedAt          time.Time `firestore:"updatedAt"`
	}
)

func planningEncode(i feature.Project) planningDocument {
	return planningDocument{SchemaVersion: 1, ID: i.ID, RepositoryID: i.SourceRepositoryID, SpaceID: i.SpaceID, SourceRepositoryID: i.SourceRepositoryID, Title: i.Title, Body: i.Body, CreatorID: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
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
	if e != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !feature.ValidID(d.ID) || (scope.SourceRepositoryID != "" && d.RepositoryID != scope.SourceRepositoryID) || (d.SpaceID != "" && d.SpaceID != scope.SpaceID) || (d.SourceRepositoryID != "" && d.SourceRepositoryID != d.RepositoryID) || c.Title != d.Title || c.Body != d.Body || d.CreatorID == "" || d.Version < 1 || d.Version > feature.MaxVersion || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return feature.Project{}, feature.ErrUnavailable
	}
	return feature.Project{ID: d.ID, SourceRepositoryID: d.RepositoryID, SpaceID: scope.SpaceID, Title: d.Title, Body: d.Body, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
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

func (s *Store) spaceParent(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("spaces").Doc(id)
}

func (s *Store) legacyParent(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(id)
}

func (s *Store) marker(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("projectStorage").Doc(id)
}

type routing struct {
	SpaceID  string `firestore:"spaceId"`
	Location string `firestore:"location"`
	Locked   bool   `firestore:"locked"`
}

func (s *Store) readParent(ctx context.Context, r feature.Scope) (*db.DocumentRef, error) {
	source := r.SourceRepositoryID
	if source == "" {
		rows, e := s.Client.Collection("channels").Doc(s.ChannelID).Collection("projectStorage").Where("spaceId", "==", r.SpaceID).Limit(2).Documents(ctx).GetAll()
		if e != nil {
			return nil, e
		}
		if len(rows) == 0 {
			return s.spaceParent(r.SpaceID), nil
		}
		if len(rows) > 1 {
			return nil, feature.ErrUnavailable
		}
		source = rows[0].Ref.ID
	}
	snap, e := s.marker(source).Get(ctx)
	if status.Code(e) == codes.NotFound {
		if r.Legacy {
			return s.legacyParent(source), nil
		}
		return s.spaceParent(r.SpaceID), nil
	}
	if e != nil {
		return nil, e
	}
	var route routing
	if e = snap.DataTo(&route); e != nil {
		return nil, e
	}
	if route.Locked || route.SpaceID != r.SpaceID {
		return nil, feature.ErrUnavailable
	}
	if route.Location == "space" {
		return s.spaceParent(r.SpaceID), nil
	}
	if route.Location != "legacy" {
		return nil, feature.ErrUnavailable
	}
	return s.legacyParent(source), nil
}

func (s *Store) planningReady() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) planningContainer(tx *db.Transaction, r feature.Scope, parent *db.DocumentRef) error {
	source := r.SourceRepositoryID
	if source != "" {
		snap, e := tx.Get(s.marker(source))
		route := routing{SpaceID: r.SpaceID, Location: "legacy"}
		if e == nil {
			if e = snap.DataTo(&route); e != nil {
				return e
			}
		} else if status.Code(e) != codes.NotFound {
			return e
		}
		if route.Locked || route.SpaceID != r.SpaceID {
			return feature.ErrUnavailable
		}
		expected := s.legacyParent(source)
		if route.Location == "space" {
			expected = s.spaceParent(r.SpaceID)
		} else if route.Location != "legacy" {
			return feature.ErrUnavailable
		}
		if parent.Path != expected.Path {
			return feature.ErrUnavailable
		}
	}
	if !r.Legacy && parent.Path != s.spaceParent(r.SpaceID).Path {
		owner, e := tx.Get(s.spaceParent(r.SpaceID))
		if e != nil {
			return e
		}
		var value struct {
			ID, GroupID, Status string
			Version             int64
		}
		if e = owner.DataTo(&value); e != nil {
			return e
		}
		if value.ID != r.SpaceID || value.GroupID != r.SpaceID || value.Status != "ACTIVE" || value.Version < 1 {
			return feature.ErrNotFound
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
	if parent.Path == s.spaceParent(r.SpaceID).Path {
		if d.ID != r.SpaceID || d.GroupID != r.SpaceID {
			return feature.ErrNotFound
		}
	} else if d.ID != source || d.GroupID != r.SpaceID {
		return feature.ErrNotFound
	} else {
		owner, e := tx.Get(s.spaceParent(r.SpaceID))
		if e == nil {
			var value struct {
				ID, GroupID, Status string
				Version             int64
			}
			if e = owner.DataTo(&value); e != nil {
				return e
			}
			if value.ID != r.SpaceID || value.GroupID != r.SpaceID || value.Status != "ACTIVE" || value.Version < 1 {
				return feature.ErrNotFound
			}
		} else if status.Code(e) != codes.NotFound {
			return e
		}
	}
	return nil
}

// A fresh source is claimed atomically only when no legacy projects or receipts exist.
func (s *Store) prepareSource(ctx context.Context, r feature.Scope) error {
	return s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		snap, e := tx.Get(s.marker(r.SourceRepositoryID))
		if e == nil {
			var v routing
			if e = snap.DataTo(&v); e != nil {
				return e
			}
			if v.Locked || v.SpaceID != r.SpaceID {
				return feature.ErrUnavailable
			}
			if v.Location == "legacy" && r.Legacy {
				return nil
			}
			if v.Location != "space" {
				return feature.ErrUnavailable
			}
			return nil
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		owner, ownerErr := tx.Get(s.spaceParent(r.SpaceID))
		if status.Code(ownerErr) == codes.NotFound && r.Legacy {
			return nil
		}
		if ownerErr != nil {
			return ownerErr
		}
		var ownerState struct {
			ID, GroupID, Status string
			Version             int64
		}
		if e := owner.DataTo(&ownerState); e != nil {
			return e
		}
		if ownerState.ID != r.SpaceID || ownerState.GroupID != r.SpaceID || ownerState.Status != "ACTIVE" || ownerState.Version < 1 {
			return feature.ErrNotFound
		}
		for _, kind := range []string{"projects", "projectsCreationReceipts"} {
			docs, e := tx.Documents(s.legacyParent(r.SourceRepositoryID).Collection(kind).Limit(1)).GetAll()
			if e != nil {
				return e
			}
			if len(docs) > 0 {
				if r.Legacy {
					return nil
				}
				return feature.ErrUnavailable
			}
		}
		return tx.Create(s.marker(r.SourceRepositoryID), map[string]any{"spaceId": r.SpaceID, "location": "space", "locked": false})
	})
}

func (s *Store) receipt(parent *db.DocumentRef, r feature.Scope, actor, key string) *db.DocumentRef {
	if parent.Path == s.spaceParent(r.SpaceID).Path {
		return parent.Collection("projectCreationScopes").Doc(r.SourceRepositoryID).Collection("receipts").Doc(planningHash([]string{actor, key}))
	}
	return parent.Collection("projectsCreationReceipts").Doc(planningHash([]string{actor, key}))
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

func (s *Store) visible(ctx context.Context, r feature.Scope, p feature.Project, parent *db.DocumentRef) error {
	if r.Legacy {
		return nil
	}
	snap, e := s.marker(p.SourceRepositoryID).Get(ctx)
	if status.Code(e) == codes.NotFound {
		return feature.ErrNotFound
	}
	if e != nil {
		return e
	}
	var route routing
	if e = snap.DataTo(&route); e != nil {
		return e
	}
	if route.Locked || route.SpaceID != r.SpaceID {
		return feature.ErrUnavailable
	}
	if route.Location == "legacy" {
		if parent.Path == s.legacyParent(p.SourceRepositoryID).Path {
			return nil
		}
		return feature.ErrNotFound
	}
	if route.Location != "space" {
		return feature.ErrUnavailable
	}
	return nil
}

func (s *Store) Get(ctx context.Context, r feature.Scope, id string) (result feature.Project, err error) {
	defer func() { err = planningError(err) }()
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	parent, e := s.readParent(ctx, r)
	if e != nil {
		return result, e
	}
	snap, e := parent.Collection("projects").Doc(id).Get(ctx)
	decodeScope := r
	if parent.Path != s.spaceParent(r.SpaceID).Path {
		decodeScope.SourceRepositoryID = parent.ID
	}
	result, err = planningDecode(snap, e, decodeScope)
	if err == nil {
		err = s.visible(ctx, r, result, parent)
	}

	return
}

func (s *Store) List(ctx context.Context, r feature.Scope, actor, token string) (result feature.Page, err error) {
	defer func() { err = planningError(err) }()
	result.Items = []feature.Project{}
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	parent, e := s.readParent(ctx, r)
	if e != nil {
		return result, e
	}
	scope := []string{s.ChannelID, r.SpaceID, actor}
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
		if parent.Path == s.spaceParent(r.SpaceID).Path {
			decodeScope.SourceRepositoryID = ""
		} else {
			decodeScope.SourceRepositoryID = parent.ID
		}
		i, e := planningDecode(snap, nil, decodeScope)
		if e != nil {
			return feature.Page{}, e
		}
		if e := s.visible(ctx, r, i, parent); e != nil {
			if errors.Is(e, feature.ErrNotFound) {
				continue
			}
			return feature.Page{}, e
		}
		if r.Legacy && i.SourceRepositoryID != r.SourceRepositoryID {
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
	if e := s.prepareSource(ctx, r); e != nil {
		return result, e
	}
	parent, e := s.readParent(ctx, r)
	if e != nil {
		return result, e
	}
	receipt := s.receipt(parent, r, actor, key)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.planningContainer(tx, r, parent); e != nil {
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
		i := feature.Project{ID: id, SourceRepositoryID: r.SourceRepositoryID, SpaceID: r.SpaceID, Title: c.Title, Body: c.Body, CreatorID: actor, Version: 1, CreatedAt: now, UpdatedAt: now}
		if e := tx.Create(parent.Collection("projects").Doc(id), planningEncode(i)); e != nil {
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
		r.SourceRepositoryID = value.SourceRepositoryID
	}
	parent, e := s.readParent(ctx, r)
	if e != nil {
		return result, e
	}
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
		return s.planningWrite(tx, ref, current, updated, actor, "update")
	})
	if err != nil {
		return result, err
	}
	return s.Get(ctx, r, id)
}
