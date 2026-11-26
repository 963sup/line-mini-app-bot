package firestore

import (
	issuedto "github.com/963sup/line-mini-app-bot/internal/business/issue/application/dto"
	repositoryapp "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	repositoryport "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
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
	OrganizationVisible func(*db.Transaction, *db.DocumentRef, string, string) (bool, error)
	CountIssues         func(func(db.Query) *db.DocumentIterator, *db.DocumentRef, string) (issuedto.StateCounts, error)
	CountDiscussions    func(*db.Transaction, *db.DocumentRef) (int64, error)
	CountProjects       func(*db.Transaction, string, string) (int64, error)
	Client              *db.Client
	ChannelID           string
}

type document struct {
	Discoverable bool             `firestore:"discoverable,omitempty"`
	ID           string           `firestore:"id"`
	OwnerType    string           `firestore:"ownerType"`
	OwnerID      string           `firestore:"ownerId"`
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
	model := repositoryDomainEntity.Repository{ID: d.ID, OwnerType: repositoryDomainEntity.OwnerType(d.OwnerType), OwnerID: d.OwnerID, Name: d.Name, Status: d.Status, Version: d.Version, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
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

func (s *Store) ObservePersonal(ctx context.Context, o repositoryDomainEntity.PersonalObservation) error {
	if !repositoryDomainEntity.ValidTypedOwner(repositoryDomainEntity.OwnerUser, o.OwnerID) || o.At.IsZero() || o.EventID == "" {
		return repositoryDomainEntity.ErrUnavailable
	}
	return s.observe(ctx, repositoryDomainEntity.OwnerUser, o.OwnerID, o.Name, o.EventID, o.At)
}

func (s *Store) observe(ctx context.Context, ownerType repositoryDomainEntity.OwnerType, ownerID, name, eventID string, at time.Time) error {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	newID, now := hex.EncodeToString(random[:]), time.Now().UTC()
	binding := s.channel().Collection("repositoryOwnershipBindings").Doc(firestorekey.JSONDigest([]string{string(ownerType), ownerID}))
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
		ref := s.repositoryRef(id)
		d := document{ID: id, OwnerType: string(ownerType), OwnerID: ownerID, Status: "ACTIVE", Name: name, Version: 1, CreatedAt: now, UpdatedAt: now}
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
			if at.Before(state.EventAt) || at.Equal(state.EventAt) && eventID == state.EventID {
				return nil
			}
		} else {
			if err := tx.Create(binding, map[string]any{"repositoryId": id, "status": "ACTIVE", "eventAt": at, "eventId": eventID}); err != nil {
				return err
			}
		}
		if bound != nil && bound.Exists() {
			if err := tx.Update(binding, []db.Update{{Path: "status", Value: "ACTIVE"}, {Path: "eventAt", Value: at}, {Path: "eventId", Value: eventID}}); err != nil {
				return err
			}
		}
		if writeRepository {
			if err := tx.Set(ref, d); err != nil {
				return err
			}
		}
		seen := map[string]bool{}
		for _, user := range []string{ownerID} {
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

type lineBindingDocument struct {
	GroupID      string    `firestore:"groupId"`
	RepositoryID string    `firestore:"repositoryId"`
	Status       string    `firestore:"status"`
	EventAt      time.Time `firestore:"eventAt"`
	EventID      string    `firestore:"eventId"`
}

func (s *Store) lineBindingRef(group string) *db.DocumentRef {
	return s.channel().Collection("repositoryLineBindings").Doc(firestorekey.Digest([]byte(group)))
}

// ObserveGroup atomically creates or advances the lifecycle of the sole LINE
// group binding. The repository lifecycle and ownership remain independent.
func (s *Store) ObserveGroup(ctx context.Context, o repositoryDomainEntity.GroupObservation) error {
	if o.At.IsZero() || o.EventID == "" || !lineid.ValidGroup(o.GroupID) || (o.Status != "ACTIVE" && o.Status != "INACTIVE") || o.Status == "ACTIVE" && !repositoryDomainEntity.ValidTypedOwner(o.OwnerType, o.OwnerID) {
		return repositoryDomainEntity.ErrUnavailable
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	newID, now := hex.EncodeToString(random[:]), time.Now().UTC()
	binding := s.lineBindingRef(o.GroupID)
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		id := newID
		snap, err := tx.Get(binding)
		if status.Code(err) == codes.NotFound {
			if o.Status == "INACTIVE" {
				return nil
			}
			if err := tx.Create(binding, lineBindingDocument{GroupID: o.GroupID, RepositoryID: id, Status: o.Status, EventAt: o.At, EventID: o.EventID}); err != nil {
				return err
			}
			if err := tx.Create(s.repositoryRef(id), document{ID: id, OwnerType: string(o.OwnerType), OwnerID: o.OwnerID, Name: o.Name, Status: "ACTIVE", Version: 1, CreatedAt: now, UpdatedAt: now}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			var current lineBindingDocument
			if err := snap.DataTo(&current); err != nil || current.GroupID != o.GroupID || !repositoryDomainEntity.ValidID(current.RepositoryID) {
				return repositoryDomainEntity.ErrUnavailable
			}
			id = current.RepositoryID
			if o.At.Before(current.EventAt) || o.At.Equal(current.EventAt) && (o.EventID == current.EventID || current.Status == "INACTIVE") {
				return nil
			}
			if err := tx.Update(binding, []db.Update{{Path: "status", Value: o.Status}, {Path: "eventAt", Value: o.At}, {Path: "eventId", Value: o.EventID}}); err != nil {
				return err
			}
		}
		if o.Status != "ACTIVE" {
			return nil
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
	if !repositoryDomainEntity.ValidID(d.ID) || d.ID != snap.Ref.ID || !repositoryDomainEntity.ValidTypedOwner(model.OwnerType, d.OwnerID) || d.Version < 1 || model.Address != nil && !model.Address.Valid() {
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

// Writable reports whether the actor can still write to this owner in the transaction.
func (s *Store) Writable(tx *db.Transaction, id, owner, actor string) (bool, error) {
	current, err := decode(tx.Get(s.repositoryRef(id)))
	if err == repositoryDomainEntity.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if current.OwnerID != owner {
		return false, nil
	}
	return repositoryapp.CanAccess(actor, current, s.organizationCheck(tx))
}

func (s *Store) organizationCheck(tx *db.Transaction) repositoryport.CheckOrganization {
	return func(org, actor string) (bool, error) {
		if s.OrganizationVisible == nil {
			return false, repositoryDomainEntity.ErrUnavailable
		}
		return s.OrganizationVisible(tx, s.channel(), org, actor)
	}
}

func (s *Store) RunAddressChange(ctx context.Context, id string, change repositoryport.AddressChange) (repositoryDomainEntity.Repository, error) {
	if !s.ready() {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrUnavailable
	}
	ref := s.repositoryRef(id)
	var updated repositoryDomainEntity.Repository
	err := s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		current, err := decode(tx.Get(ref))
		if err != nil {
			return err
		}
		updated, err = change(current, s.organizationCheck(tx))
		if err != nil {
			return err
		}
		return tx.Update(ref, []db.Update{
			{Path: "address", Value: addressDocument{CountryCode: updated.Address.CountryCode, PostalCode: updated.Address.PostalCode, AddressLine: updated.Address.AddressLine}},
			{Path: "version", Value: updated.Version},
			{Path: "updatedAt", Value: updated.UpdatedAt},
		})
	})
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	return updated, nil
}

func (s *Store) GetByGroupID(ctx context.Context, group string) (repositoryDomainEntity.Repository, error) {
	if !lineid.ValidGroup(group) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	snap, err := s.lineBindingRef(group).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	var b lineBindingDocument
	if err := snap.DataTo(&b); err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	if b.GroupID != group || b.Status != "ACTIVE" || !repositoryDomainEntity.ValidID(b.RepositoryID) {
		return repositoryDomainEntity.Repository{}, repositoryDomainEntity.ErrNotFound
	}
	repository, err := s.GetByID(ctx, b.RepositoryID)
	if err != nil {
		return repositoryDomainEntity.Repository{}, err
	}
	repository.InteractionGroupID = group
	return repository, nil
}

func (s *Store) interactionGroupID(ctx context.Context, repositoryID string) (string, error) {
	docs, err := s.channel().Collection("repositoryLineBindings").Where("repositoryId", "==", repositoryID).Limit(2).Documents(ctx).GetAll()
	if err != nil {
		return "", err
	}
	group := ""
	for _, snap := range docs {
		var binding lineBindingDocument
		if err := snap.DataTo(&binding); err != nil || binding.RepositoryID != repositoryID || !lineid.ValidGroup(binding.GroupID) {
			return "", repositoryDomainEntity.ErrUnavailable
		}
		if binding.Status != "ACTIVE" {
			continue
		}
		if group != "" {
			return "", repositoryDomainEntity.ErrUnavailable
		}
		group = binding.GroupID
	}
	return group, nil
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
		r.InteractionGroupID, err = s.interactionGroupID(ctx, r.ID)
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
