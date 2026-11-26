package domain

import (
	"errors"
	"slices"
)

type (
	Activation      struct{ Actor, TeamID, Scope, ManagementID, Maintainer string }
	ActivationState struct {
		Active                                       bool
		Scope, ManagementID, Mode, InitialMaintainer string
		Managers                                     []string
	}
)

var (
	ErrActivationDenied        = errors.New("team activation not permitted")
	ErrActivationConflict      = errors.New("team already has another management scope")
	ErrGroupMembershipRequired = errors.New("current group membership required for activation")
)

// Check returns replay=true when the requested opening has already been applied.
func (s ActivationState) Check(cmd Activation, verifiedGroupMember bool) (replay bool, err error) {
	if !s.Active || cmd.Actor == "" {
		return false, ErrActivationDenied
	}
	replay = s.Mode == "membership" && s.Scope == cmd.Scope && s.ManagementID == cmd.ManagementID && s.InitialMaintainer == cmd.Maintainer
	unbound := s.Scope == "" && s.ManagementID == "" && (s.Mode == "" || s.Mode == "legacy")
	if !slices.Contains(s.Managers, cmd.Actor) {
		if len(s.Managers) > 0 || (!replay && !unbound) {
			return false, ErrActivationDenied
		}
		if !replay && !verifiedGroupMember {
			return false, ErrGroupMembershipRequired
		}
	}
	if !replay && !unbound {
		return false, ErrActivationConflict
	}
	return replay, nil
}
