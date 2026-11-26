package orgstore

import (
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"slices"

	"cloud.google.com/go/firestore"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Run[T any](ctx context.Context, db *firestore.Client, channel string, op func(*firestore.Transaction, *firestore.DocumentRef) (*T, error)) (*T, error) {
	var result *T
	err := db.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		var e error
		result, e = op(tx, db.Collection("channels").Doc(channel))
		return e
	})
	if status.Code(err) == codes.NotFound {
		return nil, organizationDomainEntity.ErrNotFound
	}
	return result, err
}

type Document struct {
	InitialMaintainer string   `firestore:"initialMaintainerUserId"`
	Status            string   `firestore:"status"`
	MembershipMode    string   `firestore:"membershipMode"`
	Disabled          bool     `firestore:"disabled"`
	Owners            []string `firestore:"ownerUserIds"`
	Name              string   `firestore:"name"`
	Owner             string   `firestore:"ownerUserId"`
	Members           []string `firestore:"memberUserIds"`
	Enterprise        string   `firestore:"enterpriseId"`
	Managers          []string `firestore:"managerUserIds"`
	Scope             string   `firestore:"managementScope"`
	Manager           string   `firestore:"managementId"`
}

func Read(tx *firestore.Transaction, ref *firestore.DocumentRef) (Document, error) {
	snap, e := tx.Get(ref)
	if e != nil {
		return Document{}, e
	}
	var d Document
	e = snap.DataTo(&d)
	return d, e
}

func (e Document) IsOwner(user string) bool {
	return e.Owner == user || slices.Contains(e.Owners, user)
}

// LegacyOrganization translates the shared legacy document into Organization's
// own model. Enterprise owner arrays never confer Organization ownership.
func (d Document) LegacyOrganization() organizationDomainEntity.Organization {
	return organizationDomainEntity.Organization{
		Name: d.Name, EnterpriseID: d.Enterprise, OwnerUserID: d.Owner, MemberUserIDs: d.Members,
		Disabled: d.Disabled, MembershipMode: d.MembershipMode,
	}
}

func (d Document) LegacyOrganizationEnabled() bool {
	return d.LegacyOrganization().LegacyEnabled()
}

func (d Document) VisibleLegacyOrganization(user string) bool {
	return d.LegacyOrganization().VisibleLegacy(user)
}
