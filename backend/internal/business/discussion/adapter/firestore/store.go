package firestore

import (
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntityLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
)

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	discussiondto "github.com/963sup/line-mini-app-bot/internal/business/discussion/application/dto"
	discussionport "github.com/963sup/line-mini-app-bot/internal/business/discussion/application/port"
	discussionentity "github.com/963sup/line-mini-app-bot/internal/business/discussion/domain/entity"

	db "cloud.google.com/go/firestore"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/963sup/line-mini-app-bot/internal/platform/grpcstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Numberer interface {
	CreateNumberedResource(context.Context, repositoryDomainEntity.NumberedCreation, func(context.Context, *db.Transaction, repositoryDomainEntity.NumberedResource) error) (repositoryDomainEntity.NumberedResource, error)
}
type Store struct {
	Client    *db.Client
	ChannelID string
	Numbers   Numberer
	Labels    LabelValidator
}

type LabelValidator interface {
	ValidateLabelIDs(context.Context, *db.Transaction, repositoryDomainEntity.Repository, []string) error
}

var _ discussionport.Store = (*Store)(nil)

type document struct {
	SchemaVersion int64     `firestore:"schemaVersion"`
	ID            string    `firestore:"id"`
	RepositoryID  string    `firestore:"repositoryId"`
	Number        int64     `firestore:"number"`
	Title         string    `firestore:"title"`
	Body          string    `firestore:"body"`
	AuthorID      string    `firestore:"authorId"`
	Version       int64     `firestore:"version"`
	CreatedAt     time.Time `firestore:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt"`
	LabelIDs      []string  `firestore:"labelIds,omitempty"`
}

func encode(i discussionentity.Discussion) document {
	return document{1, i.ID, i.RepositoryID, i.Number, i.Title, i.Body, i.AuthorID, i.Version, i.CreatedAt, i.UpdatedAt, append([]string{}, i.LabelIDs...)}
}

func decode(snap *db.DocumentSnapshot, err error, repositoryID string) (discussionentity.Discussion, error) {
	if status.Code(err) == codes.NotFound {
		return discussionentity.Discussion{}, discussionentity.ErrNotFound
	}
	if err != nil {
		return discussionentity.Discussion{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return discussionentity.Discussion{}, err
	}
	c, err := discussionentity.Normalize(discussionentity.Content{Title: d.Title, Body: d.Body})
	ids, idErr := repositoryDomainEntityLabels.NormalizeIDs(d.LabelIDs)
	if idErr != nil {
		return discussionentity.Discussion{}, discussionport.ErrUnavailable
	}
	if err != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repositoryDomainEntity.ValidID(d.ID) || d.RepositoryID != repositoryID || d.Number < 1 || d.Number > repositoryDomainEntity.MaxResourceNumber || d.Version < 1 || d.Version > repositoryDomainEntity.MaxResourceNumber || d.AuthorID == "" || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || c.Title != d.Title || c.Body != d.Body {
		return discussionentity.Discussion{}, discussionport.ErrUnavailable
	}
	return discussionentity.Discussion{ID: d.ID, RepositoryID: d.RepositoryID, Number: d.Number, Title: d.Title, Body: d.Body, AuthorID: d.AuthorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, LabelIDs: ids}, nil
}

func (s *Store) repositoryRef(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(id)
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }

// Translate at the adapter boundary; SDK status/cause never becomes a domain dependency.

func translate(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{discussionentity.ErrInput, discussionentity.ErrNotFound, discussionentity.ErrForbidden, discussionentity.ErrConflict, discussionentity.ErrCreationConflict, discussionport.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, known) {
			return err
		}
	}
	switch {
	case errors.Is(err, repositoryDomainEntity.ErrNotFound):
		return discussionentity.ErrNotFound
	case errors.Is(err, repositoryDomainEntity.ErrCreationConflict):
		return discussionentity.ErrCreationConflict
	default:
		return grpcstatus.NormalizeContextError(err)
	}
}

func (s *Store) Get(ctx context.Context, scope repositoryDomainEntity.Repository, id string) (result discussionentity.Discussion, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, discussionport.ErrUnavailable
	}
	snap, err := s.repositoryRef(scope.ID).Collection("discussions").Doc(id).Get(ctx)
	return decode(snap, err, scope.ID)
}

type cursor struct {
	Scope  string
	Number int64
}

func (s *Store) List(ctx context.Context, scope repositoryDomainEntity.Repository, actor string, f discussiondto.Filter) (result discussiondto.Page, err error) {
	defer func() { err = translate(err) }()
	result.Items = []discussionentity.Discussion{}
	if !s.ready() {
		return result, discussionport.ErrUnavailable
	}
	key, err := json.Marshal([]string{s.ChannelID, scope.ID, actor, f.Query})
	if err != nil {
		return result, err
	}
	scopeKey := firestorekey.Digest(key)
	var c cursor
	if f.Cursor != "" {
		b, decodeErr := base64.RawURLEncoding.DecodeString(f.Cursor)
		if decodeErr != nil || json.Unmarshal(b, &c) != nil || c.Scope != scopeKey || c.Number < 1 || c.Number > repositoryDomainEntity.MaxResourceNumber {
			return result, discussionentity.ErrInput
		}
	}
	q := s.repositoryRef(scope.ID).Collection("discussions").OrderBy("number", db.Desc)
	if c.Number != 0 {
		q = q.StartAfter(c.Number)
	}
	docs, err := q.Limit(21).Documents(ctx).GetAll()
	if err != nil {
		return result, err
	}
	more := len(docs) > 20
	if more {
		docs = docs[:20]
	}
	for _, snap := range docs {
		i, err := decode(snap, nil, scope.ID)
		if err != nil {
			return discussiondto.Page{}, err
		}
		c = cursor{Scope: scopeKey, Number: i.Number}
		if strings.Contains(strings.ToLower(i.Title), strings.ToLower(f.Query)) {
			result.Items = append(result.Items, i)
		}
	}
	if more {
		b, err := json.Marshal(c)
		if err != nil {
			return discussiondto.Page{}, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func (s *Store) Create(ctx context.Context, repo repositoryDomainEntity.Repository, actorID, idempotencyKey string, content discussionentity.Content) (result discussionentity.Discussion, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() || s.Numbers == nil {
		return result, discussionport.ErrUnavailable
	}
	// This explicit v1 encoding is persisted as the creation receipt's payload hash.
	payload, err := json.Marshal(struct {
		Version int    `json:"version"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	}{1, content.Title, content.Body})
	if err != nil {
		return result, err
	}
	reference, err := s.Numbers.CreateNumberedResource(ctx, repositoryDomainEntity.NumberedCreation{RepositoryID: repo.ID, OwnerID: repo.OwnerID, ActorID: actorID, ResourceType: "discussion", IdempotencyKey: idempotencyKey, PayloadHash: firestorekey.Digest(payload)}, func(_ context.Context, tx *db.Transaction, ref repositoryDomainEntity.NumberedResource) error {
		i := discussionentity.Discussion{ID: ref.ResourceID, RepositoryID: repo.ID, Number: ref.Number, Title: content.Title, Body: content.Body, AuthorID: actorID, Version: 1, CreatedAt: ref.CreatedAt, UpdatedAt: ref.CreatedAt}
		return tx.Create(s.repositoryRef(repo.ID).Collection("discussions").Doc(i.ID), encode(i))
	})
	if err != nil {
		return result, err
	}
	// Replays return the same identity/number and its current persisted snapshot.
	return s.Get(ctx, repo, reference.ResourceID)
}

type change struct {
	ActorID string    `firestore:"actorId"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
}

func (s *Store) RunUpdate(ctx context.Context, scope repositoryDomainEntity.Repository, id string, update func(discussionport.UpdateWork) (discussionentity.Discussion, error)) (result discussionentity.Discussion, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, discussionport.ErrUnavailable
	}
	parent := s.repositoryRef(scope.ID)
	ref := parent.Collection("discussions").Doc(id)
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		var err error
		result, err = update(updateWork{ctx: ctx, tx: tx, store: s, scope: scope, parent: parent, ref: ref})
		return err
	})
	if err != nil {
		return discussionentity.Discussion{}, err
	}
	return result, nil
}

type updateWork struct {
	ctx         context.Context
	tx          *db.Transaction
	store       *Store
	scope       repositoryDomainEntity.Repository
	parent, ref *db.DocumentRef
}

func (w updateWork) Repository() (string, string, string, error) {
	snap, err := w.tx.Get(w.parent)
	if status.Code(err) == codes.NotFound {
		return "", "", "", discussionentity.ErrNotFound
	}
	if err != nil {
		return "", "", "", err
	}
	var container struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	err = snap.DataTo(&container)
	return container.ID, container.OwnerID, container.Status, err
}

func (w updateWork) Current() (discussionentity.Discussion, error) {
	snap, err := w.tx.Get(w.ref)
	return decode(snap, err, w.scope.ID)
}

func (w updateWork) ValidateLabels(ids []string) error {
	if w.store.Labels == nil {
		return discussionport.ErrUnavailable
	}
	return w.store.Labels.ValidateLabelIDs(w.ctx, w.tx, w.scope, ids)
}

func (w updateWork) Save(actor string, previous, updated discussionentity.Discussion) error {
	if err := w.tx.Set(w.ref, encode(updated)); err != nil {
		return err
	}
	return w.tx.Create(w.ref.Collection("changes").Doc(strconv.FormatInt(updated.Version, 10)), change{ActorID: actor, Version: updated.Version, At: updated.UpdatedAt})
}

func maxTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}

func CountRepositoryDiscussions(tx *db.Transaction, repository *db.DocumentRef) (int64, error) {
	q := repository.Collection("discussions").Select("id", "schemaVersion", "repositoryId")
	rows, err := tx.Documents(q).GetAll()
	if err != nil {
		return 0, err
	}
	for _, row := range rows {
		var doc document
		if err := row.DataTo(&doc); err != nil {
			return 0, err
		}
		if doc.SchemaVersion != 1 || doc.ID != row.Ref.ID || doc.RepositoryID != repository.ID {
			return 0, discussionport.ErrUnavailable
		}
	}
	return int64(len(rows)), nil
}
