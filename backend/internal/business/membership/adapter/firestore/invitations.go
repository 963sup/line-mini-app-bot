// Package membershipdb maps invitation work to one Firestore transaction.
package membershipdb

import (
	"context"
	"slices"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store struct {
	DB      *firestore.Client
	Channel string
}

func (s Store) Run(ctx context.Context, fn func(application.Work) error) error {
	root := s.DB.Collection("channels").Doc(s.Channel)
	return s.DB.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		return fn(work{tx, root})
	})
}

type work struct {
	tx   *firestore.Transaction
	root *firestore.DocumentRef
}

func (w work) organization(id string) *firestore.DocumentRef {
	return w.root.Collection("organizations").Doc(id)
}

type organizationRecord struct {
	Name         string   `firestore:"name"`
	Owner        string   `firestore:"ownerUserId"`
	Members      []string `firestore:"memberUserIds"`
	Mode         string   `firestore:"membershipMode"`
	OwnerVersion int64    `firestore:"membershipOwnerVersion"`
	Disabled     bool     `firestore:"disabled"`
}

func (w work) readOrganization(id string) (organizationRecord, error) {
	var record organizationRecord
	if uuid.Validate(id) != nil || len(id) != 36 {
		return record, application.ErrForbidden
	}
	d, err := w.tx.Get(w.organization(id))
	if status.Code(err) == codes.NotFound {
		return record, application.ErrForbidden
	}
	if err != nil {
		return record, err
	}
	if err := d.DataTo(&record); err != nil {
		return organizationRecord{}, err
	}
	if record.Owner == "" || record.Disabled || (record.Mode != "" && record.Mode != "legacy") {
		return organizationRecord{}, application.ErrForbidden
	}
	return record, nil
}

func (w work) Organization(id, user string) (string, int64, bool, error) {
	record, err := w.readOrganization(id)
	return record.Owner, record.OwnerVersion, record.Owner == user || slices.Contains(record.Members, user), err
}

func (w work) invitation(key string) *firestore.DocumentRef {
	return w.root.Collection("membershipInvitations").Doc(key)
}

func (w work) Invitation(key string) (domain.Invitation, bool, error) {
	d, err := w.tx.Get(w.invitation(key))
	if status.Code(err) == codes.NotFound {
		return domain.Invitation{}, false, nil
	}
	if err != nil {
		return domain.Invitation{}, false, err
	}
	var invitation domain.Invitation
	err = d.DataTo(&invitation)
	return invitation, err == nil, err
}

func (w work) CreateInvitation(key string, invitation domain.Invitation) error {
	return w.tx.Create(w.invitation(key), invitation)
}

func (w work) AcceptInvitation(key string, invitation domain.Invitation, at time.Time) error {
	if err := w.tx.Update(w.organization(invitation.OrganizationID), []firestore.Update{
		{Path: "memberUserIds", Value: firestore.ArrayUnion(invitation.AcceptedBy)},
	}); err != nil {
		return err
	}
	if err := w.tx.Set(w.invitation(key), invitation); err != nil {
		return err
	}
	return w.tx.Create(w.root.Collection("membershipAudits").Doc(key), map[string]any{
		"resourceType": "organization", "resourceId": invitation.OrganizationID,
		"actorUserId": invitation.AcceptedBy, "action": "accept_invitation", "at": at,
	})
}

func (w work) RevokeInvitation(key string) error {
	return w.tx.Update(w.invitation(key), []firestore.Update{{Path: "Revoked", Value: true}})
}
