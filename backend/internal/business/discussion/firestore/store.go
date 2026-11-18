package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Numberer interface {
	CreateNumberedResource(context.Context, repository.NumberedCreation, func(context.Context, *db.Transaction, repository.NumberedResource) error) (repository.NumberedResource, error)
}
type Store struct {
	Client    *db.Client
	ChannelID string
	Numbers   Numberer
	Labels    LabelValidator
}

type LabelValidator interface {
	ValidateLabelIDs(context.Context, *db.Transaction, repository.Repository, []string) error
}

var _ discussion.Store = (*Store)(nil)

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

func encode(i discussion.Discussion) document {
	return document{1, i.ID, i.RepositoryID, i.Number, i.Title, i.Body, i.AuthorID, i.Version, i.CreatedAt, i.UpdatedAt, append([]string{}, i.LabelIDs...)}
}

func decode(snap *db.DocumentSnapshot, err error, repositoryID string) (discussion.Discussion, error) {
	if status.Code(err) == codes.NotFound {
		return discussion.Discussion{}, discussion.ErrNotFound
	}
	if err != nil {
		return discussion.Discussion{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return discussion.Discussion{}, err
	}
	c, err := discussion.Normalize(discussion.Content{Title: d.Title, Body: d.Body})
	ids, idErr := labels.NormalizeIDs(d.LabelIDs)
	if idErr != nil {
		return discussion.Discussion{}, discussion.ErrUnavailable
	}
	if err != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != repositoryID || d.Number < 1 || d.Number > repository.MaxResourceNumber || d.Version < 1 || d.Version > repository.MaxResourceNumber || d.AuthorID == "" || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || c.Title != d.Title || c.Body != d.Body {
		return discussion.Discussion{}, discussion.ErrUnavailable
	}
	return discussion.Discussion{ID: d.ID, RepositoryID: d.RepositoryID, Number: d.Number, Title: d.Title, Body: d.Body, AuthorID: d.AuthorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, LabelIDs: ids}, nil
}

func (s *Store) parent(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(id)
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }

// Translate at the adapter boundary; SDK status/cause never becomes a domain dependency.
type storageFailure struct{ cause error }

func (e *storageFailure) Error() string { return "discussion storage unavailable" }
func (e *storageFailure) Unwrap() error { return discussion.ErrUnavailable }

func translate(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{discussion.ErrInput, discussion.ErrNotFound, discussion.ErrForbidden, discussion.ErrConflict, discussion.ErrCreationConflict, discussion.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, known) {
			return err
		}
	}
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return discussion.ErrNotFound
	case errors.Is(err, repository.ErrCreationConflict):
		return discussion.ErrCreationConflict
	case status.Code(err) == codes.Canceled:
		return context.Canceled
	case status.Code(err) == codes.DeadlineExceeded:
		return context.DeadlineExceeded
	default:
		return &storageFailure{cause: err}
	}
}

func (s *Store) Get(ctx context.Context, scope repository.Repository, id string) (result discussion.Discussion, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, discussion.ErrUnavailable
	}
	snap, err := s.parent(scope.ID).Collection("discussions").Doc(id).Get(ctx)
	return decode(snap, err, scope.ID)
}

type cursor struct {
	Scope  string
	Number int64
}

func hash(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }

func (s *Store) List(ctx context.Context, scope repository.Repository, actor string, f discussion.Filter) (result discussion.Page, err error) {
	defer func() { err = translate(err) }()
	result.Items = []discussion.Discussion{}
	if !s.ready() {
		return result, discussion.ErrUnavailable
	}
	key, err := json.Marshal([]string{s.ChannelID, scope.ID, actor, f.Query})
	if err != nil {
		return result, err
	}
	scopeKey := hash(key)
	var c cursor
	if f.Cursor != "" {
		b, decodeErr := base64.RawURLEncoding.DecodeString(f.Cursor)
		if decodeErr != nil || json.Unmarshal(b, &c) != nil || c.Scope != scopeKey || c.Number < 1 || c.Number > repository.MaxResourceNumber {
			return result, discussion.ErrInput
		}
	}
	q := s.parent(scope.ID).Collection("discussions").OrderBy("number", db.Desc)
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
			return discussion.Page{}, err
		}
		c = cursor{Scope: scopeKey, Number: i.Number}
		if strings.Contains(strings.ToLower(i.Title), strings.ToLower(f.Query)) {
			result.Items = append(result.Items, i)
		}
	}
	if more {
		b, err := json.Marshal(c)
		if err != nil {
			return discussion.Page{}, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func (s *Store) Create(ctx context.Context, scope repository.Repository, actor, key string, c discussion.Content) (result discussion.Discussion, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() || s.Numbers == nil {
		return result, discussion.ErrUnavailable
	}
	// This explicit v1 encoding is persisted as the creation receipt's payload hash.
	payload, err := json.Marshal(struct {
		Version int    `json:"version"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	}{1, c.Title, c.Body})
	if err != nil {
		return result, err
	}
	reference, err := s.Numbers.CreateNumberedResource(ctx, repository.NumberedCreation{RepositoryID: scope.ID, GroupID: scope.GroupID, ActorID: actor, ResourceType: "discussion", Key: key, PayloadHash: hash(payload)}, func(_ context.Context, tx *db.Transaction, ref repository.NumberedResource) error {
		i := discussion.Discussion{ID: ref.ResourceID, RepositoryID: scope.ID, Number: ref.Number, Title: c.Title, Body: c.Body, AuthorID: actor, Version: 1, CreatedAt: ref.CreatedAt, UpdatedAt: ref.CreatedAt}
		return tx.Create(s.parent(scope.ID).Collection("discussions").Doc(i.ID), encode(i))
	})
	if err != nil {
		return result, err
	}
	// Replays return the same identity/number and its current persisted snapshot.
	return s.Get(ctx, scope, reference.ResourceID)
}

type change struct {
	ActorID string    `firestore:"actorId"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
}

func (s *Store) Update(ctx context.Context, scope repository.Repository, actor, id string, input discussion.Update) (result discussion.Discussion, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, discussion.ErrUnavailable
	}
	parent := s.parent(scope.ID)
	ref := parent.Collection("discussions").Doc(id)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = discussion.Discussion{}
		snap, err := tx.Get(parent)
		if status.Code(err) == codes.NotFound {
			return discussion.ErrNotFound
		}
		if err != nil {
			return err
		}
		var container struct {
			ID      string `firestore:"id"`
			GroupID string `firestore:"groupId"`
			Status  string `firestore:"status"`
		}
		if err := snap.DataTo(&container); err != nil {
			return err
		}
		if container.ID != scope.ID || container.GroupID != scope.GroupID || container.Status != "ACTIVE" {
			return discussion.ErrNotFound
		}
		currentSnap, err := tx.Get(ref)
		current, err := decode(currentSnap, err, scope.ID)
		if err != nil {
			return err
		}
		result, err = discussion.ApplyUpdate(current, actor, input)
		if err != nil {
			return err
		}
		if result.Version == current.Version {
			return nil
		}
		if len(result.LabelIDs) > 0 {
			if s.Labels == nil {
				return discussion.ErrUnavailable
			}
			if err := s.Labels.ValidateLabelIDs(ctx, tx, scope, result.LabelIDs); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return discussion.ErrInput
				}
				return err
			}
		}
		result.UpdatedAt = maxTime(now, current.UpdatedAt)
		if err := tx.Set(ref, encode(result)); err != nil {
			return err
		}
		return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), change{ActorID: actor, Version: result.Version, At: result.UpdatedAt})
	})
	if err != nil {
		return discussion.Discussion{}, err
	}
	return result, nil
}

func maxTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}
