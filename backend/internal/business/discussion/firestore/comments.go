package firestore

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type commentDocument struct {
	SchemaVersion int64     `firestore:"schemaVersion"`
	ID            string    `firestore:"id"`
	RepositoryID  string    `firestore:"repositoryId"`
	DiscussionID  string    `firestore:"discussionId"`
	Body          string    `firestore:"body"`
	AuthorID      string    `firestore:"authorId"`
	Version       int64     `firestore:"version"`
	CreatedAt     time.Time `firestore:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt"`
}

func encodeComment(c discussion.Comment) commentDocument {
	return commentDocument{1, c.ID, c.RepositoryID, c.DiscussionID, c.Body, c.AuthorID, c.Version, c.CreatedAt, c.UpdatedAt}
}

func decodeComment(snap *db.DocumentSnapshot, err error, repositoryID, discussionID string) (discussion.Comment, error) {
	if status.Code(err) == codes.NotFound {
		return discussion.Comment{}, discussion.ErrNotFound
	}
	if err != nil {
		return discussion.Comment{}, err
	}
	var d commentDocument
	if err := snap.DataTo(&d); err != nil {
		return discussion.Comment{}, err
	}
	body, err := discussion.NormalizeComment(d.Body)
	if err != nil || body != d.Body || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != repositoryID || d.DiscussionID != discussionID || d.AuthorID == "" || d.Version < 1 || d.Version > repository.MaxResourceNumber || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return discussion.Comment{}, discussion.ErrUnavailable
	}
	return discussion.Comment{ID: d.ID, RepositoryID: d.RepositoryID, DiscussionID: d.DiscussionID, Body: d.Body, AuthorID: d.AuthorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}

func (s *Store) topic(scope repository.Repository, id string) *db.DocumentRef {
	return s.parent(scope.ID).Collection("discussions").Doc(id)
}

// Every comment write rechecks the container and topic within the same transaction.
func (s *Store) checkThread(tx *db.Transaction, scope repository.Repository, id string) error {
	snap, err := tx.Get(s.parent(scope.ID))
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
	snap, err = tx.Get(s.topic(scope, id))
	_, err = decode(snap, err, scope.ID)
	return err
}

type commentCursor struct {
	Scope, ID string
	At        time.Time
}

func (s *Store) ListComments(ctx context.Context, scope repository.Repository, actor, id, value string) (result discussion.CommentPage, err error) {
	defer func() { err = translate(err) }()
	result.Items = []discussion.Comment{}
	if !s.ready() {
		return result, discussion.ErrUnavailable
	}
	if _, err := s.Get(ctx, scope, id); err != nil {
		return result, err
	}
	key, err := json.Marshal([]string{s.ChannelID, scope.ID, actor, id})
	if err != nil {
		return result, err
	}
	scopeKey := hash(key)
	var c commentCursor
	if value != "" {
		b, e := base64.RawURLEncoding.DecodeString(value)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scopeKey || !repository.ValidID(c.ID) || c.At.IsZero() {
			return result, discussion.ErrInput
		}
	}
	q := s.topic(scope, id).Collection("comments").OrderBy("createdAt", db.Asc).OrderBy(db.DocumentID, db.Asc)
	if c.ID != "" {
		q = q.StartAfter(c.At, c.ID)
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
		item, err := decodeComment(snap, nil, scope.ID, id)
		if err != nil {
			return discussion.CommentPage{}, err
		}
		result.Items = append(result.Items, item)
		c = commentCursor{Scope: scopeKey, ID: item.ID, At: item.CreatedAt}
	}
	if more {
		b, err := json.Marshal(c)
		if err != nil {
			return discussion.CommentPage{}, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

type commentReceipt struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	PayloadHash   string `firestore:"payloadHash"`
	CommentID     string `firestore:"commentId"`
}

func (s *Store) CreateComment(ctx context.Context, scope repository.Repository, actor, id, key, body string) (result discussion.Comment, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, discussion.ErrUnavailable
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return result, err
	}
	newID, now := hex.EncodeToString(random[:]), time.Now().UTC().Truncate(time.Microsecond)
	// Explicit v1 payload encoding is independent of HTTP DTOs and server fields.
	payload, err := json.Marshal(struct {
		Version int    `json:"version"`
		Body    string `json:"body"`
	}{1, body})
	if err != nil {
		return result, err
	}
	keyScope, err := json.Marshal([]string{"discussion-comment-create", actor, key})
	if err != nil {
		return result, err
	}
	parent := s.topic(scope, id)
	receipt := parent.Collection("commentCreationReceipts").Doc(hash(keyScope))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = discussion.Comment{}
		if err := s.checkThread(tx, scope, id); err != nil {
			return err
		}
		snap, err := tx.Get(receipt)
		if err == nil {
			var prior commentReceipt
			if err := snap.DataTo(&prior); err != nil {
				return err
			}
			if prior.SchemaVersion != 1 || !repository.ValidID(prior.CommentID) {
				return discussion.ErrUnavailable
			}
			if prior.PayloadHash != hash(payload) {
				return discussion.ErrCreationConflict
			}
			snap, err := tx.Get(parent.Collection("comments").Doc(prior.CommentID))
			result, err = decodeComment(snap, err, scope.ID, id)
			if err == discussion.ErrNotFound {
				return discussion.ErrUnavailable
			}
			if err != nil {
				return err
			}
			if result.AuthorID != actor {
				return discussion.ErrUnavailable
			}
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		result = discussion.Comment{ID: newID, RepositoryID: scope.ID, DiscussionID: id, Body: body, AuthorID: actor, Version: 1, CreatedAt: now, UpdatedAt: now}
		if err := tx.Create(parent.Collection("comments").Doc(newID), encodeComment(result)); err != nil {
			return err
		}
		return tx.Create(receipt, commentReceipt{1, hash(payload), newID})
	})
	if err != nil {
		return discussion.Comment{}, err
	}
	return result, nil
}

func (s *Store) UpdateComment(ctx context.Context, scope repository.Repository, actor, id, commentID, body string, version int64) (result discussion.Comment, err error) {
	defer func() { err = translate(err) }()
	if !s.ready() {
		return result, discussion.ErrUnavailable
	}
	ref := s.topic(scope, id).Collection("comments").Doc(commentID)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = discussion.Comment{}
		if err := s.checkThread(tx, scope, id); err != nil {
			return err
		}
		snap, err := tx.Get(ref)
		current, err := decodeComment(snap, err, scope.ID, id)
		if err != nil {
			return err
		}
		result, err = discussion.ApplyCommentUpdate(current, actor, body, version)
		if err != nil {
			return err
		}
		if result.Version == current.Version {
			return nil
		}
		result.UpdatedAt = maxTime(now, current.UpdatedAt)
		if err := tx.Set(ref, encodeComment(result)); err != nil {
			return err
		}
		return tx.Create(ref.Collection("changes").Doc(strconv.FormatInt(result.Version, 10)), change{ActorID: actor, Version: result.Version, At: result.UpdatedAt})
	})
	if err != nil {
		return discussion.Comment{}, err
	}
	return result, nil
}
