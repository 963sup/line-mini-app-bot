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
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/repository/milestones"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type planningDocument struct {
	SchemaVersion int64     `firestore:"schemaVersion"`
	ID            string    `firestore:"id"`
	RepositoryID  string    `firestore:"repositoryId"`
	Title         string    `firestore:"title"`
	Body          string    `firestore:"body"`
	CreatorID     string    `firestore:"creatorId"`
	Version       int64     `firestore:"version"`
	CreatedAt     time.Time `firestore:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt"`
	State         string    `firestore:"state"`
	DueDate       string    `firestore:"dueDate"`
}

func planningEncode(i feature.Milestone) planningDocument {
	return planningDocument{SchemaVersion: 1, ID: i.ID, RepositoryID: i.RepositoryID, Title: i.Title, Body: i.Body, CreatorID: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt, State: i.State, DueDate: i.DueDate}
}

func planningDecode(snap *db.DocumentSnapshot, err error, rid string) (feature.Milestone, error) {
	if status.Code(err) == codes.NotFound {
		return feature.Milestone{}, feature.ErrNotFound
	}
	if err != nil {
		return feature.Milestone{}, err
	}
	var d planningDocument
	if e := snap.DataTo(&d); e != nil {
		return feature.Milestone{}, e
	}
	c, e := feature.Normalize(feature.Content{Title: d.Title, Body: d.Body, DueDate: d.DueDate})
	if e != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != rid || c.Title != d.Title || c.Body != d.Body || d.CreatorID == "" || d.Version < 1 || d.Version > repository.MaxResourceNumber || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || (d.State != "open" && d.State != "completed") {
		return feature.Milestone{}, feature.ErrUnavailable
	}
	return feature.Milestone{ID: d.ID, RepositoryID: d.RepositoryID, Title: d.Title, Body: d.Body, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, State: d.State, DueDate: d.DueDate}, nil
}

type planningFailure struct{ cause error }

func (e *planningFailure) Error() string { return "milestones storage unavailable" }
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

func (s *Store) planningParent(rid string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(rid)
}
func (s *Store) planningReady() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) planningContainer(tx *db.Transaction, r repository.Repository) error {
	snap, e := tx.Get(s.planningParent(r.ID))
	if status.Code(e) == codes.NotFound {
		return feature.ErrNotFound
	}
	if e != nil {
		return e
	}
	var d struct {
		ID      string `firestore:"id"`
		GroupID string `firestore:"groupId"`
		Status  string `firestore:"status"`
	}
	if e := snap.DataTo(&d); e != nil {
		return e
	}
	if d.ID != r.ID || d.GroupID != r.GroupID || d.Status != "ACTIVE" {
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
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != planningHash(scope) || !repository.ValidID(c.Last) {
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

func (s *Store) GetMilestone(ctx context.Context, r repository.Repository, id string) (result feature.Milestone, err error) {
	defer func() { err = planningError(err) }()
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	snap, e := s.planningParent(r.ID).Collection("milestones").Doc(id).Get(ctx)
	result, err = planningDecode(snap, e, r.ID)
	if err == nil {
		err = s.planningProgress(ctx, r, &result)
	}
	return
}

func (s *Store) ListMilestones(ctx context.Context, r repository.Repository, actor, token string) (result feature.Page, err error) {
	defer func() { err = planningError(err) }()
	result.Items = []feature.Milestone{}
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	scope := []string{s.ChannelID, r.ID, actor}
	q, e := planningPage(s.planningParent(r.ID).Collection("milestones").OrderBy(db.DocumentID, db.Asc), token, scope)
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
		i, e := planningDecode(snap, nil, r.ID)
		if e != nil {
			return feature.Page{}, e
		}
		if e := s.planningProgress(ctx, r, &i); e != nil {
			return feature.Page{}, e
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

func (s *Store) CreateMilestone(ctx context.Context, r repository.Repository, actor, key string, c feature.Content) (result feature.Milestone, err error) {
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
		DueDate string `json:"dueDate"`
	}{1, c.Title, c.Body, c.DueDate})
	parent := s.planningParent(r.ID)
	receipt := parent.Collection("milestonesCreationReceipts").Doc(planningHash([]string{actor, key}))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.planningContainer(tx, r); e != nil {
			return e
		}
		snap, e := tx.Get(receipt)
		if e == nil {
			var d planningReceipt
			if e := snap.DataTo(&d); e != nil {
				return e
			}
			if d.SchemaVersion != 1 || !repository.ValidID(d.ID) {
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
		i := feature.Milestone{ID: id, RepositoryID: r.ID, Title: c.Title, Body: c.Body, CreatorID: actor, Version: 1, CreatedAt: now, UpdatedAt: now, State: "open", DueDate: c.DueDate}
		if e := tx.Create(parent.Collection("milestones").Doc(id), planningEncode(i)); e != nil {
			return e
		}
		return tx.Create(receipt, planningReceipt{id, payload, 1})
	})
	if err != nil {
		return result, err
	}
	return s.GetMilestone(ctx, r, id)
}

type planningChange struct {
	ActorID string    `firestore:"actorId"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
	Action  string    `firestore:"action"`
}

func (s *Store) planningWrite(tx *db.Transaction, ref *db.DocumentRef, current, result feature.Milestone, actor, action string) error {
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

func (s *Store) UpdateMilestone(ctx context.Context, r repository.Repository, actor, id string, input feature.Update) (result feature.Milestone, err error) {
	defer func() { err = planningError(err) }()
	if !s.planningReady() {
		return result, feature.ErrUnavailable
	}
	ref := s.planningParent(r.ID).Collection("milestones").Doc(id)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.planningContainer(tx, r); e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		current, e := planningDecode(snap, e, r.ID)
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
	return s.GetMilestone(ctx, r, id)
}

func (s *Store) planningProgress(ctx context.Context, r repository.Repository, result *feature.Milestone) error {
	snaps, e := s.planningParent(r.ID).Collection("issues").Where("milestoneId", "==", result.ID).Select("state").Documents(ctx).GetAll()
	if e != nil {
		return e
	}
	result.IssueCount = int64(len(snaps))
	for _, snap := range snaps {
		state, e := snap.DataAt("state")
		if e != nil {
			return e
		}
		switch state {
		case "completed":
			result.CompletedIssueCount++
		case "open", "in_progress":
		default:
			return feature.ErrUnavailable
		}
	}
	return nil
}

// ValidateMilestoneID reads within the caller's update transaction.
func (s *Store) ValidateMilestoneID(_ context.Context, tx *db.Transaction, r repository.Repository, id string) error {
	if !s.planningReady() {
		return repository.ErrNotFound
	}
	snap, e := tx.Get(s.planningParent(r.ID).Collection("milestones").Doc(id))
	_, e = planningDecode(snap, e, r.ID)
	if errors.Is(e, feature.ErrNotFound) {
		return repository.ErrNotFound
	}
	return e
}
