// Package firestore implements session persistence using the server SDK.
package firestore

import (
	"context"
	"errors"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	"github.com/963sup/line-mini-app-bot/internal/identity/session"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct{ Client *db.Client }

var _ session.Store = (*Store)(nil)

type document struct {
	UserID            string    `firestore:"userID"`
	DisplayName       string    `firestore:"displayName"`
	ChannelID         string    `firestore:"channelID"`
	CreatedAt         time.Time `firestore:"createdAt"`
	ExpiresAt         time.Time `firestore:"expiresAt"`
	CSRFHash          string    `firestore:"csrfTokenHash"`
	AbsoluteExpiresAt time.Time `firestore:"absoluteExpiresAt,omitempty"`
	LastRenewedAt     time.Time `firestore:"lastRenewedAt,omitempty"`
	Revision          int64     `firestore:"revision,omitempty"`
	SchemaVersion     int       `firestore:"schemaVersion"`
}

func (s *Store) Replace(ctx context.Context, previous, key string, value session.Session) error {
	// Atomic replacement prevents issuing a new cookie while the old session remains active.
	// A retry only repeats database operations; credentials are generated before the transaction.
	return s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		if err := tx.Create(s.Client.Collection("sessions").Doc(key), document{
			UserID: value.User.ID, DisplayName: value.User.Name, ChannelID: value.ChannelID,
			CreatedAt: value.CreatedAt, ExpiresAt: value.ExpiresAt, CSRFHash: value.CSRFHash, SchemaVersion: 2,
			AbsoluteExpiresAt: value.AbsoluteExpiresAt, LastRenewedAt: value.LastRenewedAt, Revision: value.Revision,
		}); err != nil {
			return err
		}
		if previous != "" {
			return tx.Delete(s.Client.Collection("sessions").Doc(previous))
		}
		return nil
	})
}

func (s *Store) Get(ctx context.Context, key string) (session.Session, error) {
	snapshot, err := s.Client.Collection("sessions").Doc(key).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return session.Session{}, session.ErrInvalid
	}
	if err != nil {
		return session.Session{}, err
	}
	return decode(snapshot)
}

func decode(snapshot *db.DocumentSnapshot) (session.Session, error) {
	var doc document
	if err := snapshot.DataTo(&doc); err != nil {
		return session.Session{}, errors.New("invalid session document")
	}
	if doc.SchemaVersion != 1 && doc.SchemaVersion != 2 {
		return session.Session{}, session.ErrInvalid
	}
	if doc.SchemaVersion == 2 && (doc.AbsoluteExpiresAt.IsZero() || doc.LastRenewedAt.IsZero() || doc.Revision < 1 || doc.ExpiresAt.After(doc.AbsoluteExpiresAt)) {
		return session.Session{}, session.ErrInvalid
	}
	value := session.Session{
		User: identity.User{ID: doc.UserID, Name: doc.DisplayName}, ChannelID: doc.ChannelID,
		CreatedAt: doc.CreatedAt, ExpiresAt: doc.ExpiresAt, CSRFHash: doc.CSRFHash,
	}
	if doc.SchemaVersion == 2 {
		value.AbsoluteExpiresAt, value.LastRenewedAt, value.Revision = doc.AbsoluteExpiresAt, doc.LastRenewedAt, doc.Revision
	}
	return value, nil
}

// Renew reads before updating and never upserts, so a deleted session cannot be resurrected.
func (s *Store) Renew(ctx context.Context, key, channel, csrf string, now func() time.Time) (session.Session, error) {
	var result session.Session
	ref := s.Client.Collection("sessions").Doc(key)
	err := s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		snapshot, err := tx.Get(ref)
		if status.Code(err) == codes.NotFound {
			return session.ErrInvalid
		}
		if err != nil {
			return err
		}
		value, err := decode(snapshot)
		if err != nil {
			return err
		}
		value, changed, err := value.Renew(now(), channel, csrf)
		if err != nil {
			return err
		}
		if changed {
			if err := tx.Update(ref, []db.Update{
				{Path: "expiresAt", Value: value.ExpiresAt},
				{Path: "lastRenewedAt", Value: value.LastRenewedAt},
				{Path: "revision", Value: value.Revision},
			}); err != nil {
				return err
			}
		}
		result = value
		return nil
	})
	if err != nil {
		return session.Session{}, err
	}
	return result, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.Client.Collection("sessions").Doc(key).Delete(ctx)
	return err
}
