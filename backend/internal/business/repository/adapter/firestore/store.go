package firestore

import (
	repositoryApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestore/key"
	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"

	db "cloud.google.com/go/firestore"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	Client    *db.Client
	ChannelID string
}

type document struct {
	Discoverable bool             `firestore:"discoverable,omitempty"`
	ID           string           `firestore:"id"`
	OwnerType    string           `firestore:"ownerType"`
	OwnerID      string           `firestore:"ownerId"`
	GroupID      string           `firestore:"groupId,omitempty"`
	Name         string           `firestore:"name"`
	Status       string           `firestore:"status"`
	Version      int64            `firestore:"version"`
	CreatedAt    time.Time        `firestore:"createdAt"`
	UpdatedAt    time.Time        `firestore:"updatedAt"`
	EventAt      time.Time        `firestore:"eventAt"`
	EventID      string           `firestore:"eventId"`
	Address      *addressDocument `firestore:"address,omitempty"`
}

type addressDocument struct {
	CountryCode string `firestore:"countryCode"`
	PostalCode  string `firestore:"postalCode"`
	AddressLine string `firestore:"addressLine"`
}

func (d document) model() repositoryDomainEntity.Repository {
	model := repositoryDomainEntity.Repository{ID: d.ID, OwnerType: repositoryDomainEntity.OwnerType(d.OwnerType), OwnerID: d.OwnerID, GroupID: d.GroupID, Name: d.Name, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
	if d.Address != nil {
		model.Address = &repositoryDomainEntity.Address{CountryCode: d.Address.CountryCode, PostalCode: d.Address.PostalCode, AddressLine: d.Address.AddressLine}
	}
	return model
}

func (s *Store) ready() bool { return s != nil && s.Client != nil && s.ChannelID != "" }

func (s *Store) channel() *db.DocumentRef { return s.Client.Collection("channels").Doc(s.ChannelID) }

func (s *Store) repositoryRef(id string) *db.DocumentRef {
	return s.channel().Collection("repositories").Doc(id)
}

// Observe records either a personal owner or a LINE group binding. A group is
// an interaction binding, never the repository owner or an authorization grant.
func (s *Store) Observe(ctx context.Context, o repositoryDomainEntity.Observation) error {
	if !repositoryDomainEntity.ValidTypedOwner(o.OwnerType, o.OwnerID) || o.At.IsZero() || o.EventID == "" || (o.Status != "ACTIVE" && o.Status != "INACTIVE") || o.GroupID != "" && !lineid.ValidGroup(o.GroupID) {
		return repositoryDomainEntity.ErrUnavailable
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	newID, now := hex.EncodeToString(random[:]), time.Now().UTC()
	personal := o.GroupID == ""
	binding := s.channel().Collection("repositoryLineBindings").Doc(o.GroupID)
	if personal {
		binding = s.channel().Collection("repositoryOwnershipBindings").Doc(firestorekey.JSONDigest([]string{string(o.OwnerType), o.OwnerID}))
	}
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		id := newID
		bound, err := tx.Get(binding)
		if err == nil {
			var b struct {
				ID string `firestore:"repositoryId"`
			}
			if err := bound.DataTo(&b); err != nil {
				return err
			}
			if !repositoryDomainEntity.ValidID(b.ID) {
				return repositoryDomainEntity.ErrUnavailable
			}
			id = b.ID
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		if err != nil && status.Code(err) == codes.NotFound && o.Status == "INACTIVE" {
			return nil
		}
		ref := s.repositoryRef(id)
		d := document{ID: id, OwnerType: string(o.OwnerType), OwnerID: o.OwnerID, GroupID: o.GroupID, Status: "ACTIVE", Name: o.Name, Version: 1, CreatedAt: now, UpdatedAt: now}
		writeRepository := bound == nil || !bound.Exists()
		if bound != nil && bound.Exists() {
			var state struct {
				RepositoryID string    `firestore:"repositoryId"`
				Status       string    `firestore:"status"`
				EventAt      time.Time `firestore:"eventAt"`
				EventID      string    `firestore:"eventId"`
			}
			if err := bound.DataTo(&state); err != nil || state.RepositoryID != id {
				return repositoryDomainEntity.ErrUnavailable
			}
			if o.At.Before(state.EventAt) || o.At.Equal(state.EventAt) && (o.EventID == state.EventID || state.Status == "INACTIVE") {
				return nil
			}
		} else {
			if err := tx.Create(binding, map[string]any{"repositoryId": id, "status": o.Status, "eventAt": o.At, "eventId": o.EventID}); err != nil {
				return err
			}
		}
		if bound != nil && bound.Exists() {
			if err := tx.Update(binding, []db.Update{{Path: "status", Value: o.Status}, {Path: "eventAt", Value: o.At}, {Path: "eventId", Value: o.EventID}}); err != nil {
				return err
			}
		}
		if writeRepository {
			if err := tx.Set(ref, d); err != nil {
				return err
			}
		}
		seen := map[string]bool{}
		for _, user := range append(o.Users, o.OwnerID) {
			if user == "" || seen[user] {
				continue
			}
			seen[user] = true
			candidate := s.channel().Collection("repositoryUsers").Doc(firestorekey.Digest([]byte(user))).Collection("repositories").Doc(id)
			if err := tx.Set(candidate, map[string]any{"repositoryId": id}); err != nil {
				return err
			}
		}
		return nil
	})
}

func decode(snap *db.DocumentSnapshot, err error) (repositoryDomainEntity.Repository, error) {
	if status.Code(err) == codes.NotFound {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	model := d.model()
	if !repositoryDomainEntity.ValidID(d.ID) || d.ID != snap.Ref.ID || !repositoryDomainEntity.ValidTypedOwner(model.OwnerType, d.OwnerID) || d.GroupID != "" && !lineid.ValidGroup(d.GroupID) || d.Version < 1 || model.Address != nil && !model.Address.Valid() {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrUnavailable
	}
	return model, nil
}

func (s *Store) GetByID(ctx context.Context, id string) (repositoryDomainEntity.Repository, error) {
	if !repositoryDomainEntity.ValidID(id) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	return decode(s.repositoryRef(id).Get(ctx))
}

func (s *Store) UpdateAddress(ctx context.Context, before, after repositoryDomainEntity.Repository) (repositoryDomainEntity.Repository, error) {
	if !s.ready() || before.ID != after.ID || before.OwnerType != after.OwnerType || before.OwnerID != after.OwnerID || before.GroupID != after.GroupID || after.Address == nil || !after.Address.Valid() || after.Version != before.Version+1 {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrInput
	}
	ref := s.repositoryRef(before.ID)
	err := s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		current, err := decode(tx.Get(ref))
		if err != nil {
			return err
		}
		if current.ID != before.ID || current.OwnerType != before.OwnerType || current.OwnerID != before.OwnerID || current.GroupID != before.GroupID || current.Status != "ACTIVE" || current.Version != before.Version {
			return repositoryDomainEntity.ErrConflict
		}
		return tx.Update(ref, []db.Update{
			{Path: "address", Value: addressDocument{CountryCode: after.Address.CountryCode, PostalCode: after.Address.PostalCode, AddressLine: after.Address.AddressLine}},
			{Path: "version", Value: after.Version},
			{Path: "updatedAt", Value: after.UpdatedAt},
		})
	})
	return after, err
}

func (s *Store) GetByGroupID(ctx context.Context, group string) (repositoryDomainEntity.Repository, error) {
	if !lineid.ValidGroup(group) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	snap, err := s.channel().Collection("repositoryLineBindings").Doc(group).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	var b struct {
		ID     string `firestore:"repositoryId"`
		Status string `firestore:"status"`
	}
	if err := snap.DataTo(&b); err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	if b.Status != "ACTIVE" {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	r, err := s.GetByID(ctx, b.ID)
	if err == nil && r.GroupID != group {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrUnavailable
	}
	return r, err
}

type cursor struct{ Scope, ID string }

func (s *Store) Candidates(ctx context.Context, user, value, ownerID string) (repositoryApplicationDto.Page, error) {
	page := repositoryApplicationDto.Page{Items: []repositoryDomainEntity.Repository{}}
	scope := firestorekey.Digest([]byte(s.ChannelID + ":" + user))
	if ownerID != "" {
		scope = firestorekey.JSONDigest([]string{s.ChannelID, user, ownerID})
	}
	c := cursor{}
	if value != "" {
		b, err := base64.RawURLEncoding.DecodeString(value)
		if len(value) > 1024 || err != nil || json.Unmarshal(b, &c) != nil || c.Scope != scope || !repositoryDomainEntity.ValidID(c.ID) {
			return page, repositoryDomainEntity.ErrCursor
		}
	}
	q := s.channel().Collection("repositoryUsers").Doc(firestorekey.Digest([]byte(user))).Collection("repositories").OrderBy(db.DocumentID, db.Asc)
	if c.ID != "" {
		q = q.StartAfter(c.ID)
	}
	docs, err := q.Limit(11).Documents(ctx).GetAll()
	if err != nil {
		return page, err
	}
	if len(docs) > 10 {
		b, err := json.Marshal(cursor{Scope: scope, ID: docs[9].Ref.ID})
		if err != nil {
			return page, err
		}
		page.NextCursor = base64.RawURLEncoding.EncodeToString(b)
		docs = docs[:10]
	}
	for _, doc := range docs {
		r, err := s.GetByID(ctx, doc.Ref.ID)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, r)
	}
	return page, nil
}

// ActiveOwner checks the repository binding in the caller's write transaction.
func ActiveOwner(tx *db.Transaction, channel *db.DocumentRef, id, owner string) (bool, error) {
	snap, err := tx.Get(channel.Collection("repositories").Doc(id))
	if err != nil {
		return false, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return false, err
	}
	return d.ID == id && d.OwnerID == owner && d.Status == "ACTIVE", nil
}
