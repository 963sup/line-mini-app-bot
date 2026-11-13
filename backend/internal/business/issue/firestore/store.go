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
	"github.com/963sup/line-mini-app-bot/internal/business/issue"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Numberer interface {
	CreateNumberedResource(context.Context, repository.NumberedCreation, func(context.Context, *db.Transaction, repository.NumberedResource) error) (repository.NumberedResource, error)
}
type Store struct {
	Client     *db.Client
	ChannelID  string
	Numbers    Numberer
	Labels     LabelValidator
	Milestones MilestoneValidator
}

type MilestoneValidator interface {
	ValidateMilestoneID(context.Context, *db.Transaction, repository.Repository, string) error
}

type LabelValidator interface {
	ValidateLabelIDs(context.Context, *db.Transaction, repository.Repository, []string) error
}

var _ issue.Store = (*Store)(nil)

type document struct {
	MilestoneID   string    `firestore:"milestoneId,omitempty"`
	SchemaVersion int64     `firestore:"schemaVersion"`
	ID            string    `firestore:"id"`
	RepositoryID  string    `firestore:"repositoryId"`
	Number        int64     `firestore:"number"`
	Title         string    `firestore:"title"`
	Body          string    `firestore:"body"`
	State         string    `firestore:"state"`
	CreatorID     string    `firestore:"creatorId"`
	Version       int64     `firestore:"version"`
	CreatedAt     time.Time `firestore:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt"`
	LabelIDs      []string  `firestore:"labelIds,omitempty"`
}

func encode(i issue.Issue) document {
	return document{i.MilestoneID, 1, i.ID, i.RepositoryID, i.Number, i.Title, i.Body, i.State, i.CreatorID, i.Version, i.CreatedAt, i.UpdatedAt, append([]string{}, i.LabelIDs...)}
}

func decode(snap *db.DocumentSnapshot, err error, repositoryID string) (issue.Issue, error) {
	if status.Code(err) == codes.NotFound {
		return issue.Issue{}, issue.ErrNotFound
	}
	if err != nil {
		return issue.Issue{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return issue.Issue{}, err
	}
	c, err := issue.Normalize(issue.Content{Title: d.Title, Body: d.Body})
	ids, idErr := labels.NormalizeIDs(d.LabelIDs)
	if idErr != nil || (d.MilestoneID != "" && !repository.ValidID(d.MilestoneID)) {
		return issue.Issue{}, issue.ErrUnavailable
	}
	if err != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != repositoryID || d.Number < 1 || d.Number > repository.MaxResourceNumber || d.Version < 1 || d.Version > repository.MaxResourceNumber || !issue.ValidState(d.State) || d.CreatorID == "" || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || c.Title != d.Title || c.Body != d.Body {
		return issue.Issue{}, issue.ErrUnavailable
	}
	return issue.Issue{ID: d.ID, RepositoryID: d.RepositoryID, Number: d.Number, Title: d.Title, Body: d.Body, State: d.State, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, LabelIDs: ids, MilestoneID: d.MilestoneID}, nil
}

func (s *Store) parent(id string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(id)
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }

// Translate at the adapter boundary; SDK status/cause never becomes a domain dependency.
type storageFailure struct{ cause error }

func (e *storageFailure) Error() string { return "issue storage unavailable" }
func (e *storageFailure) Unwrap() error { return issue.ErrUnavailable }

func translate(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{issue.ErrInput, issue.ErrNotFound, issue.ErrForbidden, issue.ErrConflict, issue.ErrCreationConflict, issue.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, known) {
			return err
		}
	}
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return issue.ErrNotFound
	case errors.Is(err, repository.ErrCreationConflict):
		return issue.ErrCreationConflict
	case status.Code(err) == codes.Canceled:
		return context.Canceled
	case status.Code(err) == codes.DeadlineExceeded:
		return context.DeadlineExceeded
	default:
		return &storageFailure{cause: err}
	}
}

func (s *Store) Get(ctx context.Context, scope repository.Repository, id string) (result issue.Issue, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, issue.ErrUnavailable
	}
	snap, err := s.parent(scope.ID).Collection("issues").Doc(id).Get(ctx)
	return decode(snap, err, scope.ID)
}

type cursor struct {
	Scope  string
	Number int64
}

func hash(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }

func (s *Store) List(ctx context.Context, scope repository.Repository, actor string, f issue.Filter) (result issue.Page, err error) {
	defer func() { err = translate(err) }()
	result.Items = []issue.Issue{}
	if !s.ready() {
		return result, issue.ErrUnavailable
	}
	key, err := json.Marshal([]string{s.ChannelID, scope.ID, actor, f.State, f.Query})
	if err != nil {
		return result, err
	}
	scopeKey := hash(key)
	var c cursor
	if f.Cursor != "" {
		b, decodeErr := base64.RawURLEncoding.DecodeString(f.Cursor)
		if decodeErr != nil || json.Unmarshal(b, &c) != nil || c.Scope != scopeKey || c.Number < 1 || c.Number > repository.MaxResourceNumber {
			return result, issue.ErrInput
		}
	}
	q := s.parent(scope.ID).Collection("issues").OrderBy("number", db.Desc)
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
			return issue.Page{}, err
		}
		c = cursor{Scope: scopeKey, Number: i.Number}
		if (f.State == "" || f.State == i.State) && strings.Contains(strings.ToLower(i.Title), strings.ToLower(f.Query)) {
			result.Items = append(result.Items, i)
		}
	}
	if more {
		b, err := json.Marshal(c)
		if err != nil {
			return issue.Page{}, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func (s *Store) Create(ctx context.Context, scope repository.Repository, actor, key string, c issue.Content) (result issue.Issue, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() || s.Numbers == nil {
		return result, issue.ErrUnavailable
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
	reference, err := s.Numbers.CreateNumberedResource(ctx, repository.NumberedCreation{RepositoryID: scope.ID, GroupID: scope.GroupID, ActorID: actor, ResourceType: "issue", Key: key, PayloadHash: hash(payload)}, func(_ context.Context, tx *db.Transaction, ref repository.NumberedResource) error {
		i := issue.Issue{ID: ref.ResourceID, RepositoryID: scope.ID, Number: ref.Number, Title: c.Title, Body: c.Body, State: "open", CreatorID: actor, Version: 1, CreatedAt: ref.CreatedAt, UpdatedAt: ref.CreatedAt}
		return tx.Create(s.parent(scope.ID).Collection("issues").Doc(i.ID), encode(i))
	})
	if err != nil {
		return result, err
	}
	// Replays return the same identity/number and its current persisted snapshot.
	return s.Get(ctx, scope, reference.ResourceID)
}

type change struct {
	ActorID string    `firestore:"actorId"`
	From    string    `firestore:"fromState"`
	To      string    `firestore:"toState"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
}

func (s *Store) Update(ctx context.Context, scope repository.Repository, actor, id string, input issue.Update) (result issue.Issue, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, issue.ErrUnavailable
	}
	parent := s.parent(scope.ID)
	ref := parent.Collection("issues").Doc(id)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = issue.Issue{}
		snap, err := tx.Get(parent)
		if status.Code(err) == codes.NotFound {
			return issue.ErrNotFound
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
			return issue.ErrNotFound
		}
		currentSnap, err := tx.Get(ref)
		current, err := decode(currentSnap, err, scope.ID)
		if err != nil {
			return err
		}
		result, err = issue.ApplyUpdate(current, actor, input)
		if err != nil {
			return err
		}
		if result.Version == current.Version {
			return nil
		}
		if len(result.LabelIDs) > 0 {
			if s.Labels == nil {
				return issue.ErrUnavailable
			}
			if err := s.Labels.ValidateLabelIDs(ctx, tx, scope, result.LabelIDs); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return issue.ErrInput
				}
				return err
			}
		}
		if result.MilestoneID != "" {
			if s.Milestones == nil {
				return issue.ErrUnavailable
			}
			if e := s.Milestones.ValidateMilestoneID(ctx, tx, scope, result.MilestoneID); e != nil {
				if errors.Is(e, repository.ErrNotFound) {
					return issue.ErrInput
				}
				return e
			}
		}
		result.UpdatedAt = maxTime(now, current.UpdatedAt)
		if err := tx.Set(ref, encode(result)); err != nil {
			return err
		}
		return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), change{actor, current.State, result.State, result.Version, result.UpdatedAt})
	})
	if err != nil {
		return issue.Issue{}, err
	}
	return result, nil
}

func maxTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}
