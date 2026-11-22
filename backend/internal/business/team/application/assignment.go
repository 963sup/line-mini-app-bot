package application

import (
	"context"
	"errors"

	"github.com/963sup/line-mini-app-bot/internal/business/team/domain"
)

var ErrAssignmentNotFound = errors.New("team assignment scope not found")

type AssignmentWork interface {
	Scope() (domain.AssignmentScope, error)
	Authority(enterpriseID, actor string) (domain.AssignmentAuthority, error)
	Save(assigned bool) error
}

type AssignmentStore interface {
	RunAssignment(ctx context.Context, organizationID, teamID string, apply func(AssignmentWork) error) error
}

func AssignEnterpriseTeam(ctx context.Context, store AssignmentStore, actor, organizationID, teamID string, assigned bool) error {
	return store.RunAssignment(ctx, organizationID, teamID, func(work AssignmentWork) error {
		scope, err := work.Scope()
		if err != nil {
			return err
		}
		if !scope.AllowsAssignment() {
			return ErrAssignmentNotFound
		}
		authority, err := work.Authority(scope.EnterpriseID, actor)
		if err != nil {
			return err
		}
		if !authority.AllowsAssignment() {
			return ErrAssignmentNotFound
		}
		return work.Save(assigned)
	})
}
