package firestore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	feature "github.com/963sup/line-mini-app-bot/internal/business/repository/milestones"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
)

type milestoneDocument struct {
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

func encodeMilestone(i feature.Milestone) milestoneDocument {
	return milestoneDocument{SchemaVersion: 1, ID: i.ID, RepositoryID: i.RepositoryID, Title: i.Title, Body: i.Body, CreatorID: i.CreatorID, Version: i.Version, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt, State: i.State, DueDate: i.DueDate}
}

func decodeMilestone(snap *db.DocumentSnapshot, err error, rid string) (feature.Milestone, error) {
	if status.Code(err) == codes.NotFound {
		return feature.Milestone{}, feature.ErrNotFound
	}
	if err != nil {
		return feature.Milestone{}, err
	}
	var d milestoneDocument
	if e := snap.DataTo(&d); e != nil {
		return feature.Milestone{}, e
	}
	c, e := feature.Normalize(feature.Content{Title: d.Title, Body: d.Body, DueDate: d.DueDate})
	if e != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != rid || c.Title != d.Title || c.Body != d.Body || d.CreatorID == "" || d.Version < 1 || d.Version > repository.MaxResourceNumber || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || (d.State != "open" && d.State != "completed") {
		return feature.Milestone{}, feature.ErrUnavailable
	}
	return feature.Milestone{ID: d.ID, RepositoryID: d.RepositoryID, Title: d.Title, Body: d.Body, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, State: d.State, DueDate: d.DueDate}, nil
}

func (s *Store) checkActiveRepository(tx *db.Transaction, r repository.Repository) error {
	snap, e := tx.Get(s.repositoryRef(r.ID))
	if status.Code(e) == codes.NotFound {
		return feature.ErrNotFound
	}
	if e != nil {
		return e
	}
	var d struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	if e := snap.DataTo(&d); e != nil {
		return e
	}
	if !r.MatchesActive(d.ID, d.OwnerID, d.Status) {
		return feature.ErrNotFound
	}
	return nil
}

func applyMilestoneCursor(q db.Query, token string, scope any) (db.Query, error) {
	if token != "" {
		last, ok := firestorekey.DecodeCursor(token, scope)
		if !ok || !repository.ValidID(last) {
			return q, feature.ErrInput
		}
		q = q.StartAfter(last)
	}
	return q.Limit(21), nil
}

func (s *Store) GetMilestone(ctx context.Context, r repository.Repository, id string) (result feature.Milestone, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	snap, e := s.repositoryRef(r.ID).Collection("milestones").Doc(id).Get(ctx)
	result, err = decodeMilestone(snap, e, r.ID)
	if err == nil {
		err = s.loadMilestoneIssueCounts(ctx, r, &result)
	}
	return
}

func (s *Store) ListMilestones(ctx context.Context, r repository.Repository, actor, token string) (result feature.Page, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	result.Items = []feature.Milestone{}
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	scope := []string{s.ChannelID, r.ID, actor}
	q, e := applyMilestoneCursor(s.repositoryRef(r.ID).Collection("milestones").OrderBy(db.DocumentID, db.Asc), token, scope)
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
		i, e := decodeMilestone(snap, nil, r.ID)
		if e != nil {
			return feature.Page{}, e
		}
		if e := s.loadMilestoneIssueCounts(ctx, r, &i); e != nil {
			return feature.Page{}, e
		}
		result.Items = append(result.Items, i)
	}
	if more {
		result.NextCursor = firestorekey.EncodeCursor(scope, snaps[len(snaps)-1].Ref.ID)
	}
	return result, nil
}

type milestoneReceipt struct {
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
	SchemaVersion int64  `firestore:"schemaVersion"`
}

func (s *Store) CreateMilestone(ctx context.Context, r repository.Repository, actor, key string, c feature.Content) (result feature.Milestone, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
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
		DueDate string `json:"dueDate"`
	}{1, c.Title, c.Body, c.DueDate})
	parent := s.repositoryRef(r.ID)
	receipt := parent.Collection("milestonesCreationReceipts").Doc(firestorekey.JSONDigest([]string{actor, key}))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.checkActiveRepository(tx, r); e != nil {
			return e
		}
		snap, e := tx.Get(receipt)
		if e == nil {
			var d milestoneReceipt
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
		if e := tx.Create(parent.Collection("milestones").Doc(id), encodeMilestone(i)); e != nil {
			return e
		}
		return tx.Create(receipt, milestoneReceipt{id, payload, 1})
	})
	if err != nil {
		return result, err
	}
	return s.GetMilestone(ctx, r, id)
}

type milestoneChange struct {
	ActorID string    `firestore:"actorId"`
	Version int64     `firestore:"version"`
	At      time.Time `firestore:"at"`
	Action  string    `firestore:"action"`
}

func writeMilestoneChange(tx *db.Transaction, ref *db.DocumentRef, current, result feature.Milestone, actor, action string) error {
	if result.Version == current.Version {
		return nil
	}
	result.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if result.UpdatedAt.Before(current.UpdatedAt) {
		result.UpdatedAt = current.UpdatedAt
	}
	if e := tx.Set(ref, encodeMilestone(result)); e != nil {
		return e
	}
	return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), milestoneChange{actor, result.Version, result.UpdatedAt, action})
}

func (s *Store) UpdateMilestone(ctx context.Context, r repository.Repository, actor, id string, input feature.Update) (result feature.Milestone, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, feature.ErrUnavailable
	}
	ref := s.repositoryRef(r.ID).Collection("milestones").Doc(id)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if e := s.checkActiveRepository(tx, r); e != nil {
			return e
		}
		snap, e := tx.Get(ref)
		current, e := decodeMilestone(snap, e, r.ID)
		if e != nil {
			return e
		}
		updated, e := feature.ApplyUpdate(current, actor, input)
		if e != nil {
			return e
		}
		return writeMilestoneChange(tx, ref, current, updated, actor, "update")
	})
	if err != nil {
		return result, err
	}
	return s.GetMilestone(ctx, r, id)
}

func (s *Store) loadMilestoneIssueCounts(ctx context.Context, r repository.Repository, result *feature.Milestone) error {
	snaps, e := s.repositoryRef(r.ID).Collection("issues").Where("milestoneId", "==", result.ID).Select("state").Documents(ctx).GetAll()
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
	if !s.ready() {
		return repository.ErrNotFound
	}
	snap, e := tx.Get(s.repositoryRef(r.ID).Collection("milestones").Doc(id))
	_, e = decodeMilestone(snap, e, r.ID)
	if errors.Is(e, feature.ErrNotFound) {
		return repository.ErrNotFound
	}
	return e
}
