package application

import (
	"context"
	"slices"

	"github.com/963sup/line-mini-app-bot/internal/platform/line/lineid"

	membership "github.com/963sup/line-mini-app-bot/internal/business/membership/domain"
	"github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

// OwnershipWork exposes the facts and write that must share one transaction.
type OwnershipWork interface {
	Load() (entity.Organization, []string, bool, error)
	SaveOwner(previous, successor string) error
}

type OwnershipStore interface {
	RunOwnershipTransfer(ctx context.Context, organizationID string, transfer func(OwnershipWork) (*entity.Organization, error)) (*entity.Organization, error)
}

func TransferOwnership(ctx context.Context, store OwnershipStore, actor, organizationID, target string) (*entity.Organization, error) {
	return store.RunOwnershipTransfer(ctx, organizationID, func(work OwnershipWork) (*entity.Organization, error) {
		org, members, legacyEnabled, err := work.Load()
		if err != nil {
			return nil, err
		}
		if org.OwnerUserID != actor || !legacyEnabled {
			return nil, entity.ErrNotFound
		}
		successor := target
		if !lineid.ValidUser(successor) {
			successor = ""
			for _, member := range members {
				if lineid.ValidUser(member) && membership.MemberRef(organizationID, member) == target {
					successor = member
					break
				}
			}
		}
		result, err := org.TransferOwnership(actor, successor, slices.Contains(members, successor))
		if err != nil {
			return nil, err
		}
		if successor == actor {
			return &result, nil
		}
		return &result, work.SaveOwner(actor, successor)
	})
}
