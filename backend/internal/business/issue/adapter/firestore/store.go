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

	db "cloud.google.com/go/firestore"
	issuedto "github.com/963sup/line-mini-app-bot/internal/business/issue/application/dto"
	issueport "github.com/963sup/line-mini-app-bot/internal/business/issue/application/port"
	issueentity "github.com/963sup/line-mini-app-bot/internal/business/issue/domain/entity"
	issuepolicy "github.com/963sup/line-mini-app-bot/internal/business/issue/domain/policy"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/963sup/line-mini-app-bot/internal/platform/grpcstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Numberer interface {
	CreateNumberedResource(context.Context, repositoryDomainEntity.NumberedCreation, func(context.Context, *db.Transaction, repositoryDomainEntity.NumberedResource) error) (repositoryDomainEntity.NumberedResource, error)
}
type Store struct {
	RepositoryWritable func(*db.Transaction, string, string, string) (bool, error)
	Client             *db.Client
	ChannelID          string
	Numbers            Numberer
	Labels             LabelValidator
	Milestones         MilestoneValidator
}

type MilestoneValidator interface {
	ValidateMilestoneID(context.Context, *db.Transaction, repositoryDomainEntity.Repository, string) error
}

type LabelValidator interface {
	ValidateLabelIDs(context.Context, *db.Transaction, repositoryDomainEntity.Repository, []string) error
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
	ids, idErr := repositoryDomainEntityLabels.NormalizeIDs(d.LabelIDs)
	if idErr != nil || (d.MilestoneID != "" && !repositoryDomainEntity.ValidID(d.MilestoneID)) {
		return issueentity.Issue{}, issueport.ErrUnavailable
	}
	if err != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repositoryDomainEntity.ValidID(d.ID) || d.RepositoryID != repositoryID || d.Number < 1 || d.Number > repositoryDomainEntity.MaxResourceNumber || d.Version < 1 || d.Version > repositoryDomainEntity.MaxResourceNumber || !issuepolicy.ValidState(d.State) || d.CreatorID == "" || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || c.Title != d.Title || c.Body != d.Body {
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
	case errors.Is(err, repositoryDomainEntity.ErrNotFound):
		return issueentity.ErrNotFound
	case errors.Is(err, repositoryDomainEntity.ErrCreationConflict):
		return issueentity.ErrCreationConflict
	default:
		return grpcstatus.NormalizeContextError(err)
	}
}

func (s *Store) Get(ctx context.Context, scope repositoryDomainEntity.Repository, id string) (result issueentity.Issue, err error) {
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

func (s *Store) List(ctx context.Context, scope repositoryDomainEntity.Repository, actor string, f issuedto.Filter) (result issuedto.Page, err error) {
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
		if decodeErr != nil || json.Unmarshal(b, &c) != nil || c.Scope != scopeKey || c.Number < 1 || c.Number > repositoryDomainEntity.MaxResourceNumber {
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

func (s *Store) Create(ctx context.Context, repo repositoryDomainEntity.Repository, actorID, idempotencyKey string, c issueentity.Content) (result issueentity.Issue, err error) {
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
	reference, err := s.Numbers.CreateNumberedResource(ctx, repositoryDomainEntity.NumberedCreation{RepositoryID: repo.ID, OwnerID: repo.OwnerID, ActorID: actorID, ResourceType: "issue", IdempotencyKey: idempotencyKey, PayloadHash: firestorekey.Digest(payload)}, func(_ context.Context, tx *db.Transaction, ref repositoryDomainEntity.NumberedResource) error {
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

func (s *Store) RunUpdate(ctx context.Context, scope repositoryDomainEntity.Repository, id string, update func(issueport.UpdateWork) (issueentity.Issue, error)) (result issueentity.Issue, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, issueport.ErrUnavailable
	}
	parent := s.repositoryRef(scope.ID)
	ref := parent.Collection("issues").Doc(id)
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		var err error
		result, err = update(updateWork{ctx: ctx, tx: tx, store: s, scope: scope, parent: parent, ref: ref})
		return err
	})
	if err != nil {
		return issueentity.Issue{}, err
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

func (w updateWork) CheckRepository(actor string) error {
	return w.store.checkWrite(w.tx, w.scope, actor)
}

func (s *Store) checkWrite(tx *db.Transaction, r repositoryDomainEntity.Repository, actor string) error {
	if s.RepositoryWritable == nil {
		return issueport.ErrUnavailable
	}
	allowed, err := s.RepositoryWritable(tx, r.ID, r.OwnerID, actor)
	if err != nil {
		return err
	}
	if !allowed {
		return issueentity.ErrNotFound
	}
	return nil
}

func (w updateWork) Current() (issueentity.Issue, error) {
	snap, err := w.tx.Get(w.ref)
	return decode(snap, err, w.scope.ID)
}

func (w updateWork) ValidateLabels(ids []string) error {
	if w.store.Labels == nil {
		return issueport.ErrUnavailable
	}
	return w.store.Labels.ValidateLabelIDs(w.ctx, w.tx, w.scope, ids)
}

func (w updateWork) ValidateMilestone(id string) error {
	if w.store.Milestones == nil {
		return issueport.ErrUnavailable
	}
	return w.store.Milestones.ValidateMilestoneID(w.ctx, w.tx, w.scope, id)
}

func (w updateWork) Save(actor string, previous, updated issueentity.Issue) error {
	if err := w.tx.Set(w.ref, encode(updated)); err != nil {
		return err
	}
	return w.tx.Create(w.ref.Collection("changes").Doc(strconv.FormatInt(updated.Version, 10)), change{actor, previous.State, updated.State, updated.Version, updated.UpdatedAt})
}

func CountRepositoryIssues(documents func(db.Query) *db.DocumentIterator, repository *db.DocumentRef, milestoneID string) (issuedto.StateCounts, error) {
	q := repository.Collection("issues").Select("id", "schemaVersion", "repositoryId", "state")
	if milestoneID != "" {
		q = q.Where("milestoneId", "==", milestoneID)
	}
	rows, err := documents(q).GetAll()
	if err != nil {
		return issuedto.StateCounts{}, err
	}
	var counts issuedto.StateCounts
	for _, row := range rows {
		var doc document
		if err := row.DataTo(&doc); err != nil {
			return issuedto.StateCounts{}, err
		}
		if doc.SchemaVersion != 1 || doc.ID != row.Ref.ID || doc.RepositoryID != repository.ID {
			return issuedto.StateCounts{}, issueport.ErrUnavailable
		}
		switch doc.State {
		case "open":
			counts.Open++
		case "in_progress":
			counts.InProgress++
		case "completed":
			counts.Completed++
		default:
			return issuedto.StateCounts{}, issueport.ErrUnavailable
		}
	}
	return counts, nil
}
