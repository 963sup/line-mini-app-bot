package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/project/application"

	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"

	projectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/project/application/dto"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

type referenceStore struct {
	store
	hydrate bool
	changes int
}

func (s *referenceStore) ListReferences(_ context.Context, _ projectApplicationPort.Scope, _, _, _ string, hydrate bool) (projectApplicationDto.ReferencePage, error) {
	s.hydrate = hydrate
	return projectApplicationDto.ReferencePage{Items: []projectDomainEntity.Reference{{SourceType: "object", SourceUnavailable: !hydrate}}}, nil
}

func (s *referenceStore) ChangeReference(context.Context, projectApplicationPort.Scope, string, string, projectDomainEntity.Reference, bool, int64) (projectDomainEntity.Project, error) {
	s.changes++
	return s.item, nil
}

func (s *referenceStore) ReferencePresent(context.Context, projectApplicationPort.Scope, string, string, projectDomainEntity.Reference) (bool, error) {
	return true, nil
}

func TestReferenceIdentityAndIndependentAuthorization(t *testing.T) {
	id := strings.Repeat("a", 32)
	for _, kind := range []string{"object", "issue", "discussion"} {
		if !projectDomainEntity.ValidReference(projectDomainEntity.Reference{SourceType: kind, SourceID: id}) {
			t.Fatal(kind)
		}
	}
	if projectDomainEntity.ValidReference(projectDomainEntity.Reference{SourceType: "subobject", SourceID: id}) {
		t.Fatal("child binding")
	}
	gid := "C" + id
	st := &referenceStore{store: store{item: projectDomainEntity.Project{ID: id, OwnerID: gid, DefaultRepositoryID: id, CreatorID: "owner", Version: 1}}}
	src := &sources{denied: true, space: gid}
	svc := &application.Service{Groups: &groups{}, Sources: src, Store: st}
	page, e := svc.ListReferences(context.Background(), "owner", gid, id, "")
	if e != nil || st.hydrate || !page.Items[0].SourceUnavailable {
		t.Fatal(page, e)
	}
	if _, e = svc.AddReference(context.Background(), "owner", gid, id, projectDomainEntity.Reference{SourceType: "object", SourceID: id}, 1); !errors.Is(e, projectDomainEntity.ErrNotFound) || st.changes != 0 {
		t.Fatal("source denial bypass", e)
	}
	if _, e = svc.RemoveReference(context.Background(), "member", gid, id, id, 1); !errors.Is(e, projectDomainEntity.ErrForbidden) || st.changes != 0 {
		t.Fatal("owner denial bypass", e)
	}
	if _, e = svc.RemoveReference(context.Background(), "owner", gid, id, id, 1); e != nil || st.changes != 1 {
		t.Fatal("revoked reference trapped", e)
	}
	present, matches, e := svc.ReferenceState(context.Background(), "owner", gid, id, projectDomainEntity.Reference{SourceRepositoryID: id, SourceType: "object", SourceID: id}, true)
	if e != nil || !present || !matches {
		t.Fatal("source permission must not block identity-only check", e)
	}
	_, matches, e = svc.ReferenceState(context.Background(), "owner", gid, id, projectDomainEntity.Reference{ID: id}, false)
	if e != nil || matches {
		t.Fatal("present reference cannot confirm absence", e)
	}
	if _, _, e = svc.ReferenceState(context.Background(), "member", gid, id, projectDomainEntity.Reference{ID: id}, true); !errors.Is(e, projectDomainEntity.ErrForbidden) {
		t.Fatal("non-creator state check", e)
	}
	if _, _, e = svc.ReferenceState(context.Background(), "owner", gid, id, projectDomainEntity.Reference{ID: id, SourceID: id}, true); !errors.Is(e, projectDomainEntity.ErrInput) {
		t.Fatal("ambiguous identity", e)
	}
}
