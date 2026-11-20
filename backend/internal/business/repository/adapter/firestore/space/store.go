package firestore

import (
	repositoryApplicationDtoSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto/space"
	repositoryApplicationPortSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port/space"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	repositoryDomainEntitySpace "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/space"
)

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	db "cloud.google.com/go/firestore"

	"github.com/963sup/line-mini-app-bot/internal/persistence/grpcstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/963sup/line-mini-app-bot/internal/persistence/firestorekey"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

var _ repositoryApplicationPortSpace.Store = (*Store)(nil)

type document struct {
	ID            string    `firestore:"id"`
	RepositoryID  string    `firestore:"repositoryId"`
	Name          string    `firestore:"name"`
	CreatorID     string    `firestore:"creatorId"`
	SchemaVersion int64     `firestore:"schemaVersion"`
	Version       int64     `firestore:"version"`
	CreatedAt     time.Time `firestore:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt"`
}

func encode(s repositoryDomainEntitySpace.Space) document {
	return document{ID: s.ID, RepositoryID: s.RepositoryID, Name: s.Name, CreatorID: s.CreatorID, SchemaVersion: 1, Version: s.Version, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

func decode(snap *db.DocumentSnapshot, err error, rid string) (repositoryDomainEntitySpace.Space, error) {
	if status.Code(err) == codes.NotFound {
		return repositoryDomainEntitySpace.Space{}, repositoryDomainEntitySpace.ErrNotFound
	}
	if err != nil {
		return repositoryDomainEntitySpace.Space{}, err
	}
	var d document
	if e := snap.DataTo(&d); e != nil {
		return repositoryDomainEntitySpace.Space{}, e
	}
	n, e := repositoryDomainEntitySpace.NormalizeName(d.Name)
	if e != nil || n != d.Name || d.ID != snap.Ref.ID || !repositoryDomainEntity.ValidID(d.ID) || d.RepositoryID != rid || d.SchemaVersion != 1 || d.CreatorID == "" || d.Version != 1 || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) {
		return repositoryDomainEntitySpace.Space{}, repositoryDomainEntitySpace.ErrUnavailable
	}
	return repositoryDomainEntitySpace.Space{ID: d.ID, RepositoryID: d.RepositoryID, Name: d.Name, CreatorID: d.CreatorID, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, nil
}
func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }
func (s *Store) repositoryRef(rid string) *db.DocumentRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("repositories").Doc(rid)
}

func (s *Store) Get(ctx context.Context, r repositoryDomainEntity.Repository, id string) (result repositoryDomainEntitySpace.Space, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, repositoryDomainEntitySpace.ErrUnavailable
	}
	snap, e := s.repositoryRef(r.ID).Collection("spaces").Doc(id).Get(ctx)
	return decode(snap, e, r.ID)
}

type cursor struct{ Scope, ID string }

func (s *Store) List(ctx context.Context, r repositoryDomainEntity.Repository, actor, token string) (result repositoryApplicationDtoSpace.Page, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	result.Items = []repositoryDomainEntitySpace.Space{}
	if !s.ready() {
		return result, repositoryDomainEntitySpace.ErrUnavailable
	}
	scope := firestorekey.JSONDigest([]string{s.ChannelID, r.ID, actor})
	c := cursor{}
	if token != "" {
		b, e := base64.RawURLEncoding.DecodeString(token)
		if e != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repositoryDomainEntity.ValidID(c.ID) {
			return result, repositoryDomainEntitySpace.ErrInput
		}
	}
	q := s.repositoryRef(r.ID).Collection("spaces").OrderBy(db.DocumentID, db.Asc)
	if c.ID != "" {
		q = q.StartAfter(c.ID)
	}
	docs, e := q.Limit(21).Documents(ctx).GetAll()
	if e != nil {
		return result, e
	}
	more := len(docs) > 20
	if more {
		docs = docs[:20]
	}
	for _, snap := range docs {
		i, e := decode(snap, nil, r.ID)
		if e != nil {
			return repositoryApplicationDtoSpace.Page{}, e
		}
		result.Items = append(result.Items, i)
		c = cursor{scope, i.ID}
	}
	if more {
		b, _ := json.Marshal(c)
		result.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return result, nil
}

func container(tx *db.Transaction, parent *db.DocumentRef, r repositoryDomainEntity.Repository) error {
	snap, e := tx.Get(parent)
	if status.Code(e) == codes.NotFound {
		return repositoryDomainEntitySpace.ErrNotFound
	}
	if e != nil {
		return e
	}
	var d struct {
		ID      string `firestore:"id"`
		OwnerID string `firestore:"ownerId"`
		Status  string `firestore:"status"`
	}
	if e = snap.DataTo(&d); e != nil {
		return e
	}
	if !r.MatchesActive(d.ID, d.OwnerID, d.Status) {
		return repositoryDomainEntitySpace.ErrNotFound
	}
	return nil
}

type receipt struct {
	SchemaVersion int64  `firestore:"schemaVersion"`
	ID            string `firestore:"id"`
	PayloadHash   string `firestore:"payloadHash"`
}

func (s *Store) Create(ctx context.Context, r repositoryDomainEntity.Repository, actor, key string, name string) (result repositoryDomainEntitySpace.Space, err error) {
	defer func() { err = grpcstatus.NormalizeContextError(err) }()
	if !s.ready() {
		return result, repositoryDomainEntitySpace.ErrUnavailable
	}
	var random [16]byte
	if _, e := rand.Read(random[:]); e != nil {
		return result, e
	}
	id := hex.EncodeToString(random[:])
	now := time.Now().UTC().Truncate(time.Microsecond)
	parent := s.repositoryRef(r.ID)
	ref := parent.Collection("spaces").Doc(id)
	// Space receipts do not allocate repository resource numbers.
	rec := parent.Collection("spaceCreationReceipts").Doc(firestorekey.JSONDigest([]string{actor, key}))
	payload := firestorekey.JSONDigest(struct{ Name string }{name})
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		result = repositoryDomainEntitySpace.Space{}
		if e := container(tx, parent, r); e != nil {
			return e
		}
		prior, e := tx.Get(rec)
		if e == nil {
			var old receipt
			if e = prior.DataTo(&old); e != nil {
				return e
			}
			if old.SchemaVersion != 1 || !repositoryDomainEntity.ValidID(old.ID) || len(old.PayloadHash) != 64 {
				return repositoryDomainEntitySpace.ErrUnavailable
			}
			if old.PayloadHash != payload {
				return repositoryDomainEntitySpace.ErrCreationConflict
			}
			snap, e := tx.Get(parent.Collection("spaces").Doc(old.ID))
			result, e = decode(snap, e, r.ID)
			if e == nil && result.CreatorID != actor {
				return repositoryDomainEntitySpace.ErrUnavailable
			}
			return e
		}
		if status.Code(e) != codes.NotFound {
			return e
		}
		result = repositoryDomainEntitySpace.Space{ID: id, RepositoryID: r.ID, Name: name, CreatorID: actor, Version: 1, CreatedAt: now, UpdatedAt: now}
		if e = tx.Create(ref, encode(result)); e != nil {
			return e
		}
		return tx.Create(rec, receipt{1, id, payload})
	})
	if err != nil {
		return repositoryDomainEntitySpace.Space{}, err
	}
	return result, nil
}
