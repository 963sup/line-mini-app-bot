package command

import (
	"context"
	"errors"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/enterprise/application/port"
	enterprise "github.com/963sup/line-mini-app-bot/internal/business/enterprise/domain"
)

type commandStore struct {
	port.Store
	enterprise port.EnterpriseFacts
	attachErr  error
	writes     int
}

func (s *commandStore) RunOwnerChange(_ context.Context, _, _ string, apply func(port.OwnerChange) (*dto.Link, error)) (*dto.Link, error) {
	return apply(commandWork{s})
}

func (s *commandStore) RunOrganizationLink(_ context.Context, _, _ string, apply func(port.OrganizationLink) (*dto.Link, error)) (*dto.Link, error) {
	return apply(commandWork{s})
}

type commandWork struct{ store *commandStore }

func (w commandWork) Enterprise(string) (port.EnterpriseFacts, error) { return w.store.enterprise, nil }

func (w commandWork) AttachForOwner(actor string) error {
	if actor != "actor" {
		return errors.New("wrong actor")
	}
	if w.store.attachErr != nil {
		return w.store.attachErr
	}
	w.store.writes++
	return nil
}
func (w commandWork) AddOwner() error { w.store.writes++; return nil }

func TestEnterpriseCommandsOwnErrorsAndRejectWithoutWrites(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		attach, active, owner bool
		attachErr             error
		want                  error
	}{
		{name: "inactive owner change", owner: true, want: enterprise.ErrNotFound},
		{name: "non-owner change", active: true, want: enterprise.ErrNotFound},
		{name: "owner change", active: true, owner: true},
		{name: "inactive enterprise link", attach: true, owner: true, want: enterprise.ErrNotFound},
		{attachErr: enterprise.ErrNotFound, name: "inactive organization link", attach: true, active: true, owner: true, want: enterprise.ErrNotFound},
		{name: "foreign enterprise", attach: true, active: true, owner: true, attachErr: enterprise.ErrConflict, want: enterprise.ErrConflict},
		{name: "attach organization", attach: true, active: true, owner: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &commandStore{enterprise: port.EnterpriseFacts{Name: "HQ", Active: tc.active, ActorIsOwner: tc.owner}, attachErr: tc.attachErr}
			var value *dto.Link
			var err error
			if tc.attach {
				value, err = AttachOrganization(t.Context(), store, "actor", "enterprise", "organization")
			} else {
				value, err = AddOwner(t.Context(), store, "actor", "enterprise", "successor")
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if tc.want != nil {
				if value != nil || store.writes != 0 {
					t.Fatal("denied operation exposed data or wrote state")
				}
			} else if value == nil || value.ID != "enterprise" || value.Name != "HQ" || store.writes != 1 {
				t.Fatal("successful operation did not commit its own result")
			}
		})
	}
}
