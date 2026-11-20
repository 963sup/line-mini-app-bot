package firestore

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	db "cloud.google.com/go/firestore"
	issuedto "github.com/963sup/line-mini-app-bot/internal/business/issue/application/dto"
	issueport "github.com/963sup/line-mini-app-bot/internal/business/issue/application/port"
	issueentity "github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
	issuepolicy "github.com/963sup/line-mini-app-bot/internal/business/issue/domain/policy"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"
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

var _ issueport.Store = (*Store)(nil)

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

func encode(i issueentity.Issue) document {
	return document{i.MilestoneID, 1, i.ID, i.RepositoryID, i.Number, i.Title, i.Body, i.State, i.CreatorID, i.Version, i.CreatedAt, i.UpdatedAt, append([]string{}, i.LabelIDs...)}
}

func decode(snap *db.DocumentSnapshot, err error, repositoryID string) (issueentity.Issue, error) {
	if status.Code(err) == codes.NotFound {
		return issueentity.Issue{}, issueentity.ErrNotFound
	}
	if err != nil {
		return issueentity.Issue{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return issueentity.Issue{}, err
	}
	c, err := issueentity.Normalize(issueentity.Content{Title: d.Title, Body: d.Body})
	ids, idErr := labels.NormalizeIDs(d.LabelIDs)
	if idErr != nil || (d.MilestoneID != "" && !repository.ValidID(d.MilestoneID)) {
		return issueentity.Issue{}, issueport.ErrUnavailable
	}
	if err != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != repositoryID || d.Number < 1 || d.Number > repository.MaxResourceNumber || d.Version < 1 || d.Version > repository.MaxResourceNumber || !issuepolicy.ValidState(d.State) || d.CreatorID == "" || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || c.Title != d.Title || c.Body != d.Body {
		return issueentity.Issue{}, issueport.ErrUnavailable
	}
	return issueentity.Issue{ID: d.ID, RepositoryID: d.RepositoryID, Number: d.Number, Title: d.Title, Body: d.Body, State: d.State, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, LabelIDs: ids, MilestoneID: d.MilestoneID}, nil
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
	for _, known := range []error{issueentity.ErrInput, issueentity.ErrNotFound, issueentity.ErrForbidden, issueentity.ErrConflict, issueentity.ErrCreationConflict, issueport.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, known) {
			return err
		}
	}
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return issueentity.ErrNotFound
	case errors.Is(err, repository.ErrCreationConflict):
		return issueentity.ErrCreationConflict
	default:
		return grpcstatus.NormalizeContextError(err)
	}
}

func (s *Store) Get(ctx context.Context, scope repository.Repository, id string) (result issueentity.Issue, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, issueport.ErrUnavailable
	}
	snap, err := s.repositoryRef(scope.ID).Collection("issues").Doc(id).Get(ctx)
	return decode(snap, err, scope.ID)
}

type cursor struct {
	Scope  string
	Number int64
}

func (s *Store) List(ctx context.Context, scope repository.Repository, actor string, f issuedto.Filter) (result issuedto.Page, err error) {
	defer func() { err = translate(err) }()
	result.Items = []issueentity.Issue{}
	if !s.ready() {
		return result, issueport.ErrUnavailable
	}
	key, err := json.Marshal([]string{s.ChannelID, scope.ID, actor, f.State, f.Query})
	if err != nil {
		return result, err
	}
	scopeKey := firestorekey.Digest(key)
	var c cursor
	if f.Cursor != "" {
		b, decodeErr := base64.RawURLEncoding.DecodeString(f.Cursor)
		if decodeErr != nil || json.Unmarshal(b, &c) != nil || c.Scope != scopeKey || c.Number < 1 || c.Number > repository.MaxResourceNumber {
			return result, issueentity.ErrInput
		}
	}
	q := s.repositoryRef(scope.ID).Collection("issues").OrderBy("number", db.Desc)
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
			return issuedto.Page{}, err
		}
		c = cursor{Scope: scopeKey, Number: i.Number}
		if (f.State == "" || f.State == i.State) && strings.Contains(strings.ToLower(i.Title), strings.ToLower(f.Query)) {
			result.Items = append(result.Items, i)
		}
	}
	if more {
		b, err := json.Marshal(c)
		if err != nil {
			return issuedto.Page{}, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func (s *Store) Create(ctx context.Context, repo repository.Repository, actorID, idempotencyKey string, c issueentity.Content) (result issueentity.Issue, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() || s.Numbers == nil {
		return result, issueport.ErrUnavailable
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
	reference, err := s.Numbers.CreateNumberedResource(ctx, repository.NumberedCreation{RepositoryID: repo.ID, OwnerID: repo.OwnerID, ActorID: actorID, ResourceType: "issue", IdempotencyKey: idempotencyKey, PayloadHash: firestorekey.Digest(payload)}, func(_ context.Context, tx *db.Transaction, ref repository.NumberedResource) error {
		i := issueentity.Issue{ID: ref.ResourceID, RepositoryID: repo.ID, Number: ref.Number, Title: c.Title, Body: c.Body, State: "open", CreatorID: actorID, Version: 1, CreatedAt: ref.CreatedAt, UpdatedAt: ref.CreatedAt}
		return tx.Create(s.repositoryRef(repo.ID).Collection("issues").Doc(i.ID), encode(i))
	})
	if err != nil {
		return result, err
	}
	// Replays return the same identity/number and its current persisted snapshot.
	return s.Get(ctx, repo, reference.ResourceID)
}

type change struct {
	ActorID string    `firestore:"actorId"`
	From    string    `firestore:"fromState"`
	To      string    `firestore:"toState"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
}

func (s *Store) Update(ctx context.Context, scope repository.Repository, actor, id string, input issueentity.Update) (result issueentity.Issue, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, issueport.ErrUnavailable
	}
	parent := s.repositoryRef(scope.ID)
	ref := parent.Collection("issues").Doc(id)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		result = issueentity.Issue{}
		snap, err := tx.Get(parent)
		if status.Code(err) == codes.NotFound {
			return issueentity.ErrNotFound
		}
		if err != nil {
			return err
		}
		var container struct {
			ID      string `firestore:"id"`
			OwnerID string `firestore:"ownerId"`
			Status  string `firestore:"status"`
		}
		if err := snap.DataTo(&container); err != nil {
			return err
		}
		if !scope.MatchesActive(container.ID, container.OwnerID, container.Status) {
			return issueentity.ErrNotFound
		}
		currentSnap, err := tx.Get(ref)
		current, err := decode(currentSnap, err, scope.ID)
		if err != nil {
			return err
		}
		result, err = issueentity.ApplyUpdate(current, actor, input)
		if err != nil {
			return err
		}
		if result.Version == current.Version {
			return nil
		}
		if len(result.LabelIDs) > 0 {
			if s.Labels == nil {
				return issueport.ErrUnavailable
			}
			if err := s.Labels.ValidateLabelIDs(ctx, tx, scope, result.LabelIDs); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return issueentity.ErrInput
				}
				return err
			}
		}
		if result.MilestoneID != "" {
			if s.Milestones == nil {
				return issueport.ErrUnavailable
			}
			if e := s.Milestones.ValidateMilestoneID(ctx, tx, scope, result.MilestoneID); e != nil {
				if errors.Is(e, repository.ErrNotFound) {
					return issueentity.ErrInput
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
		return issueentity.Issue{}, err
	}
	return result, nil
}

func maxTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}
