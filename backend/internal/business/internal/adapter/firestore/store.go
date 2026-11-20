package orgstore

import (
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

import (
	"context"
	"regexp"
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
	Owners     []string `firestore:"ownerUserIds"`
	Name       string   `firestore:"name"`
	Owner      string   `firestore:"ownerUserId"`
	Members    []string `firestore:"memberUserIds"`
	Enterprise string   `firestore:"enterpriseId"`
	Managers   []string `firestore:"managerUserIds"`
	Scope      string   `firestore:"managementScope"`
	Manager    string   `firestore:"managementId"`
}

var UserID = regexp.MustCompile("^U[0-9a-f]{32}$")

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
