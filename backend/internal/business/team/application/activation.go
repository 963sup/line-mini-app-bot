package application

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/team/domain"
)

type (
	ActivationStore interface {
		CommitActivation(context.Context, domain.Activation, bool) error
	}
	MembershipCheck func(context.Context, string, string) (bool, error)
)

func Activate(ctx context.Context, store ActivationStore, check MembershipCheck, cmd domain.Activation) error {
	err := store.CommitActivation(ctx, cmd, false)
	if !errors.Is(err, domain.ErrGroupMembershipRequired) {
		return err
	}
	if check == nil {
		return errors.New("group membership verification unavailable")
	}
	member, err := check(ctx, cmd.TeamID, cmd.Actor)
	if err != nil {
		return err
	}
	if !member {
		return domain.ErrActivationDenied
	}
	return store.CommitActivation(ctx, cmd, true)
}
