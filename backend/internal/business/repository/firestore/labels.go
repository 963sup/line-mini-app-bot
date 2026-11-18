package firestore

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ labels.Store = (*Store)(nil)

// ValidateLabelIDs only reads within the caller's transaction, before its writes.
func (s *Store) ValidateLabelIDs(_ context.Context, tx *db.Transaction, scope repository.Repository, ids []string) error {
	if !s.ready() {
		return labels.ErrUnavailable
	}
	if len(ids) > 10 {
		return repository.ErrNotFound
	}
	for _, id := range ids {
		if !repository.ValidID(id) {
			return repository.ErrNotFound
		}
		snap, err := tx.Get(s.repositoryRef(scope.ID).Collection("labels").Doc(id))
		if status.Code(err) == codes.NotFound {
			return repository.ErrNotFound
		}
		if err != nil {
			return err
		}
		if _, err := decodeLabel(snap, nil, scope.ID); err != nil {
			return err
		}
	}
	return nil
}

type labelDocument struct {
	SchemaVersion int64     `firestore:"schemaVersion"`
	ID            string    `firestore:"id"`
	RepositoryID  string    `firestore:"repositoryId"`
	Name          string    `firestore:"name"`
	NameKey       string    `firestore:"nameKey"`
	AuthorID      string    `firestore:"authorId"`
	CreatedAt     time.Time `firestore:"createdAt"`
}
type (
	labelBinding struct {
		SchemaVersion int64  `firestore:"schemaVersion"`
		LabelID       string `firestore:"labelId"`
	}
	labelReceipt struct {
		SchemaVersion int64  `firestore:"schemaVersion"`
		LabelID       string `firestore:"labelId"`
		PayloadHash   string `firestore:"payloadHash"`
	}
)

func labelError(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{labels.ErrInput, labels.ErrNotFound, labels.ErrNameConflict, labels.ErrCreationConflict, labels.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, known) {
			return known
		}
	}
	if errors.Is(err, repository.ErrNotFound) {
		return labels.ErrNotFound
	}
	return grpcstatus.NormalizeContextError(err)
}

func decodeLabel(snap *db.DocumentSnapshot, err error, scope string) (labels.Label, error) {
	if status.Code(err) == codes.NotFound {
		return labels.Label{}, labels.ErrUnavailable
	}
	if err != nil {
		return labels.Label{}, err
	}
	var d labelDocument
	if err := snap.DataTo(&d); err != nil {
		return labels.Label{}, err
	}
	name, err := labels.NormalizeName(d.Name)
	if err != nil || d.SchemaVersion != 1 || d.ID != snap.Ref.ID || !repository.ValidID(d.ID) || d.RepositoryID != scope || name != d.Name || d.NameKey != labels.NameKey(d.Name) || d.AuthorID == "" || d.CreatedAt.IsZero() {
		return labels.Label{}, labels.ErrUnavailable
	}
	return labels.Label{ID: d.ID, RepositoryID: d.RepositoryID, Name: d.Name, CreatedAt: d.CreatedAt}, nil
}

func (s *Store) ListLabels(ctx context.Context, scope repository.Repository, actor, value string) (result labels.Page, err error) {
	defer func() { err = labelError(err) }()
	result.Items = []labels.Label{}
	if !s.ready() {
		return result, labels.ErrUnavailable
	}
	encoded, err := json.Marshal([]string{"repository-label-list", s.ChannelID, scope.ID, actor})
	if err != nil {
		return result, err
	}
	scopeKey := firestorekey.Digest(encoded)
	var c cursor
	if value != "" {
		b, e := base64.RawURLEncoding.DecodeString(value)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scopeKey || !repository.ValidID(c.ID) {
			return result, labels.ErrInput
		}
	}
	q := s.repositoryRef(scope.ID).Collection("labels").OrderBy(db.DocumentID, db.Asc)
	if c.ID != "" {
		q = q.StartAfter(c.ID)
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
		item, err := decodeLabel(snap, nil, scope.ID)
		if err != nil {
			return labels.Page{}, err
		}
		result.Items = append(result.Items, item)
		c = cursor{Scope: scopeKey, ID: item.ID}
	}
	if more {
		b, err := json.Marshal(c)
		if err != nil {
			return labels.Page{}, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func (s *Store) CreateLabel(ctx context.Context, scope repository.Repository, actor, creationKey, name string) (result labels.Label, err error) {
	defer func() { err = labelError(err) }()
	if !s.ready() {
		return result, labels.ErrUnavailable
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return result, err
	}
	id, now := hex.EncodeToString(random[:]), time.Now().UTC().Truncate(time.Microsecond)
	payload, err := json.Marshal(struct {
		Version int    `json:"version"`
		Name    string `json:"name"`
	}{1, name})
	if err != nil {
		return result, err
	}
	encoded, err := json.Marshal([]string{"repository-label-create", actor, creationKey})
	if err != nil {
		return result, err
	}
	parent := s.repositoryRef(scope.ID)
	catalog := parent.Collection("labels")
	receipt := parent.Collection("labelCreationReceipts").Doc(firestorekey.Digest(encoded))
	nameKey := labels.NameKey(name)
	binding := parent.Collection("labelNames").Doc(firestorekey.Digest([]byte(nameKey)))
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = labels.Label{}
		container, err := decode(tx.Get(parent))
		if err != nil {
			return err
		}
		if container.OwnerID != scope.OwnerID || container.Status != "ACTIVE" {
			return labels.ErrNotFound
		}
		snap, err := tx.Get(receipt)
		if err == nil {
			var prior labelReceipt
			if err := snap.DataTo(&prior); err != nil {
				return err
			}
			if prior.SchemaVersion != 1 || !repository.ValidID(prior.LabelID) {
				return labels.ErrUnavailable
			}
			if prior.PayloadHash != firestorekey.Digest(payload) {
				return labels.ErrCreationConflict
			}
			item, err := tx.Get(catalog.Doc(prior.LabelID))
			result, err = decodeLabel(item, err, scope.ID)
			if err != nil {
				return err
			}
			if result.Name != name {
				return labels.ErrUnavailable
			}
			snap, err = tx.Get(binding)
			if status.Code(err) == codes.NotFound {
				return labels.ErrUnavailable
			}
			if err != nil {
				return err
			}
			var linked labelBinding
			if err := snap.DataTo(&linked); err != nil {
				return err
			}
			if linked.SchemaVersion != 1 || linked.LabelID != result.ID {
				return labels.ErrUnavailable
			}
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		snap, err = tx.Get(binding)
		if err == nil {
			var linked labelBinding
			if err := snap.DataTo(&linked); err != nil {
				return err
			}
			if linked.SchemaVersion != 1 || !repository.ValidID(linked.LabelID) {
				return labels.ErrUnavailable
			}
			snap, err := tx.Get(catalog.Doc(linked.LabelID))
			existing, err := decodeLabel(snap, err, scope.ID)
			if err != nil {
				return err
			}
			if labels.NameKey(existing.Name) != nameKey {
				return labels.ErrUnavailable
			}
			return labels.ErrNameConflict
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		d := labelDocument{1, id, scope.ID, name, nameKey, actor, now}
		if err := tx.Create(catalog.Doc(id), d); err != nil {
			return err
		}
		if err := tx.Create(binding, labelBinding{1, id}); err != nil {
			return err
		}
		if err := tx.Create(receipt, labelReceipt{1, id, firestorekey.Digest(payload)}); err != nil {
			return err
		}
		result = labels.Label{ID: id, RepositoryID: scope.ID, Name: name, CreatedAt: now}
		return nil
	})
	if err != nil {
		return labels.Label{}, err
	}
	return result, nil
}
