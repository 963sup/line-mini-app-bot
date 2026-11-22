package query_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	groupdto "github.com/963sup/line-mini-app-bot/internal/linegroup/application/dto"
	groupquery "github.com/963sup/line-mini-app-bot/internal/linegroup/application/query"
	groupentity "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/entity"
	groupevent "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/event"
	groupvalueobject "github.com/963sup/line-mini-app-bot/internal/linegroup/domain/valueobject"
)

type store struct {
	current groupentity.Group
	missing bool
}

func (s *store) Observe(context.Context, groupevent.Observation) error { return nil }
func (s *store) GetByID(context.Context, string) (groupentity.Group, error) {
	if s.missing {
		return groupentity.Group{}, groupentity.ErrNotFound
	}
	return s.current, nil
}

func (s *store) Candidates(context.Context, string, string) (groupdto.Page, error) {
	return groupdto.Page{Items: []groupentity.Group{s.current}}, nil
}

type member struct {
	allowed bool
	calls   int
}

func (m *member) IsMember(context.Context, string, string) (bool, error) {
	m.calls++
	return m.allowed, nil
}

func TestGroupMustExistAndMemberMustBeCurrent(t *testing.T) {
	ctx := context.Background()
	id := "C" + strings.Repeat("1", 32)
	cases := []struct {
		name, status           string
		missing, allowed, want bool
	}{
		{"missing group", "", true, true, false},
		{"missing outsider denied", "", true, false, false},
		{"active member", "ACTIVE", false, true, true},
		{"active outsider denied", "ACTIVE", false, false, false},
		{"existing inactive never falls back", "INACTIVE", false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &store{missing: tc.missing, current: groupentity.Group{ID: id, GroupID: id, Status: tc.status}}
			m := &member{allowed: tc.allowed}
			svc := &groupquery.Service{Store: s, Membership: m}
			got, err := svc.Get(ctx, "actor", id)
			if tc.want {
				if err != nil || got.ID != id {
					t.Fatalf("expected member Group, got %+v %v", got, err)
				}
			} else if !errors.Is(err, groupentity.ErrNotFound) {
				t.Fatalf("must deny rather than infer legacy permission: %+v %v", got, err)
			}
			if tc.status == "INACTIVE" && m.calls != 0 {
				t.Fatal("inactive Group must fail before membership fallback")
			}
		})
	}
}

func TestSpaceCandidatesAreNotAuthorization(t *testing.T) {
	id := "C" + strings.Repeat("1", 32)
	s := &store{current: groupentity.Group{ID: id, GroupID: id, Status: "ACTIVE"}}
	m := &member{allowed: false}
	svc := &groupquery.Service{Store: s, Membership: m}
	page, err := svc.List(context.Background(), "actor", "")
	if err != nil || len(page.Items) != 0 || m.calls != 1 {
		t.Fatalf("candidate must be rechecked: %+v %v calls=%d", page, err, m.calls)
	}
}

func TestLegacyGroupContextDoesNotAcceptBusinessGroupID(t *testing.T) {
	groupID := "C" + strings.Repeat("1", 32)
	businessGroupID := strings.Repeat("2", 32)
	if !groupvalueobject.ValidID(groupID) || groupvalueobject.ValidID(businessGroupID) {
		t.Fatal("legacy /groups identifiers must remain LINE GroupIDs")
	}
	m := &member{allowed: true}
	svc := &groupquery.Service{Store: &store{missing: true}, Membership: m}
	if _, err := svc.Get(context.Background(), "actor", businessGroupID); !errors.Is(err, groupentity.ErrNotFound) {
		t.Fatalf("business GroupID must not become a group authorization context: %v", err)
	}
	if m.calls != 0 {
		t.Fatal("invalid group context must not reach group membership")
	}
}

type teamQualificationFunc func(context.Context, string, string) (bool, error)

func (f teamQualificationFunc) IsTeamMember(ctx context.Context, id, actor string) (bool, error) {
	return f(ctx, id, actor)
}

func TestResourceGroupRequiresFormalTeamMembership(t *testing.T) {
	id := "C" + strings.Repeat("a", 32)
	for _, allowed := range []bool{false, true} {
		s := groupquery.Service{Teams: teamQualificationFunc(func(_ context.Context, teamID, actor string) (bool, error) {
			if teamID != id || actor != "actor" {
				t.Fatalf("unexpected authority scope: %q %q", teamID, actor)
			}
			return allowed, nil
		})}
		got, err := s.Get(context.Background(), "actor", id)
		if allowed && (err != nil || got.ID != id) {
			t.Fatalf("active membership did not authorize: %+v %v", got, err)
		}
		if !allowed && !errors.Is(err, groupentity.ErrNotFound) {
			t.Fatalf("missing membership must deny without LINE fallback: %v", err)
		}
	}
}
