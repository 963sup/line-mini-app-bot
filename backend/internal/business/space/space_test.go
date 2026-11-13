package space_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	sp "github.com/963sup/line-mini-app-bot/internal/business/space"
)

type store struct {
	current sp.Space
	missing bool
}

func (s *store) Observe(context.Context, sp.Observation) error { return nil }
func (s *store) GetByID(context.Context, string) (sp.Space, error) {
	if s.missing {
		return sp.Space{}, sp.ErrNotFound
	}
	return s.current, nil
}

func (s *store) GetByGroupID(ctx context.Context, id string) (sp.Space, error) {
	return s.GetByID(ctx, id)
}

func (s *store) Candidates(context.Context, string, string) (sp.Page, error) {
	return sp.Page{Items: []sp.Space{s.current}}, nil
}

type member struct {
	allowed bool
	calls   int
}

func (m *member) IsMember(context.Context, string, string) (bool, error) {
	m.calls++
	return m.allowed, nil
}

func TestLegacyGetOnlyAllowsMissingRecordTransition(t *testing.T) {
	ctx := context.Background()
	id := "C" + strings.Repeat("1", 32)
	cases := []struct {
		name, status           string
		missing, allowed, want bool
	}{
		{"missing member transition", "", true, true, true},
		{"missing outsider denied", "", true, false, false},
		{"active member", "ACTIVE", false, true, true},
		{"active outsider denied", "ACTIVE", false, false, false},
		{"existing inactive never falls back", "INACTIVE", false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &store{missing: tc.missing, current: sp.Space{ID: id, GroupID: id, Status: tc.status}}
			m := &member{allowed: tc.allowed}
			svc := &sp.Service{Store: s, Membership: m}
			got, err := svc.LegacyGet(ctx, "actor", id)
			if tc.want {
				if err != nil || got.ID != id {
					t.Fatalf("expected member Space, got %+v %v", got, err)
				}
			} else if !errors.Is(err, sp.ErrNotFound) {
				t.Fatalf("must deny rather than infer legacy permission: %+v %v", got, err)
			}
			if tc.status == "INACTIVE" && m.calls != 0 {
				t.Fatal("inactive Space must fail before membership fallback")
			}
		})
	}
}

func TestSpaceCandidatesAreNotAuthorization(t *testing.T) {
	id := "C" + strings.Repeat("1", 32)
	s := &store{current: sp.Space{ID: id, GroupID: id, Status: "ACTIVE"}}
	m := &member{allowed: false}
	svc := &sp.Service{Store: s, Membership: m}
	page, err := svc.List(context.Background(), "actor", "")
	if err != nil || len(page.Items) != 0 || m.calls != 1 {
		t.Fatalf("candidate must be rechecked: %+v %v calls=%d", page, err, m.calls)
	}
}
