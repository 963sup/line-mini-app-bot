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

type groups struct{ denied bool }

func (s *groups) AuthorizeGroup(_ context.Context, actor, id string) (string, error) {
	if s.denied || actor == "outsider" {
		return "", projectDomainEntity.ErrNotFound
	}
	return id, nil
}

type sources struct {
	denied bool
	space  string
}

func (s *sources) Authorize(_ context.Context, a, id, owner string) (projectDomainEntity.Source, error) {
	if s.denied || a == "outsider" || owner != "" && owner != s.space {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrNotFound
	}
	return projectDomainEntity.Source{ID: id, OwnerID: s.space}, nil
}

type store struct {
	item     projectDomainEntity.Project
	progress int
}

func (s *store) List(context.Context, projectApplicationPort.Scope, string, string) (projectApplicationDto.Page, error) {
	return projectApplicationDto.Page{Items: []projectDomainEntity.Project{s.item}}, nil
}

func (s *store) Get(context.Context, projectApplicationPort.Scope, string) (projectDomainEntity.Project, error) {
	return s.item, nil
}

func (s *store) Create(_ context.Context, r projectApplicationPort.Scope, a, k string, c projectDomainEntity.Content) (projectDomainEntity.Project, error) {
	s.item = projectDomainEntity.Project{ID: strings.Repeat("2", 32), OwnerID: r.OwnerID, DefaultRepositoryID: c.SourceRepositoryID, CreatorID: a, Title: c.Title, Body: c.Body, Version: 1}
	return s.item, nil
}

func (s *store) RunUpdate(_ context.Context, _ projectApplicationPort.Scope, _ string, update func(projectApplicationPort.UpdateWork) error) (projectDomainEntity.Project, error) {
	if err := update(metadataWork{store: s}); err != nil {
		return projectDomainEntity.Project{}, err
	}
	return s.item, nil
}

type metadataWork struct {
	projectApplicationPort.UpdateWork
	store *store
}

func (w metadataWork) Current() (projectDomainEntity.Project, error) { return w.store.item, nil }

func (w metadataWork) Save(_ string, _, updated projectDomainEntity.Project, _ projectApplicationPort.UpdateWrites) error {
	w.store.item = updated
	return nil
}

func (s *store) ListItems(context.Context, projectApplicationPort.Scope, string, string, string) (projectDomainEntity.ItemPage, error) {
	return projectDomainEntity.ItemPage{}, nil
}

func (s *store) ChangeItem(context.Context, projectApplicationPort.Scope, string, string, string, bool, int64) (projectDomainEntity.Project, error) {
	return s.item, nil
}

func (s *store) Progress(context.Context, projectApplicationPort.Scope, *projectDomainEntity.Project) error {
	s.progress++
	return nil
}

func TestGroupProjectMetadataSurvivesSourceDenial(t *testing.T) {
	ctx := context.Background()
	gid := "C" + strings.Repeat("1", 32)
	rid := strings.Repeat("3", 32)
	pid := strings.Repeat("2", 32)
	ss := &groups{}
	src := &sources{space: gid}
	st := &store{item: projectDomainEntity.Project{ID: pid, OwnerID: gid, DefaultRepositoryID: rid, CreatorID: "owner", Title: "own metadata", Version: 1}}
	svc := &application.Service{Groups: ss, Sources: src, Store: st}
	if _, err := svc.Create(ctx, "owner", "C"+strings.Repeat("5", 32), "cross-space-key-0001", projectDomainEntity.Content{Title: "must reject foreign source", SourceRepositoryID: rid}); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatalf("Group membership must not authorize another Group source: %v", err)
	}
	src.denied = true
	got, e := svc.Get(ctx, "owner", gid, pid)
	if e != nil || got.Title != "own metadata" || got.SourceAvailable || st.progress != 0 {
		t.Fatalf("own metadata must survive source denial without progress: %+v %v calls=%d", got, e, st.progress)
	}
	page, e := svc.List(ctx, "owner", gid, "")
	if e != nil || len(page.Items) != 1 || page.Items[0].SourceAvailable {
		t.Fatalf("Group list must remain own metadata: %+v %v", page, e)
	}
	if _, e = svc.ListItems(ctx, "owner", gid, pid, ""); !errors.Is(e, projectDomainEntity.ErrNotFound) {
		t.Fatalf("source item read must deny after source revoke: %v", e)
	}
	if _, e = svc.ChangeItem(ctx, "owner", gid, pid, strings.Repeat("4", 32), true, 1); !errors.Is(e, projectDomainEntity.ErrNotFound) {
		t.Fatalf("source item mutation must deny after revoke: %v", e)
	}
	_, e = svc.Update(ctx, "owner", gid, pid, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: "edited own metadata"}, ExpectedVersion: 1})
	if e != nil {
		t.Fatalf("own metadata update must not require source: %v", e)
	}
	if _, e = svc.Update(ctx, "member", gid, pid, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: "forbidden"}, ExpectedVersion: 2}); !errors.Is(e, projectDomainEntity.ErrForbidden) {
		t.Fatalf("member cannot manage another Project: %v", e)
	}
	ss.denied = true
	src.denied = false
	if _, e = svc.List(ctx, "owner", gid, ""); !errors.Is(e, projectDomainEntity.ErrNotFound) {
		t.Fatalf("canonical must deny Group revoke: %v", e)
	}
	if _, e = svc.List(ctx, "owner", rid, ""); !errors.Is(e, projectDomainEntity.ErrNotFound) {
		t.Fatalf("legacy cannot bypass Group revoke using source auth: %v", e)
	}
}

type progressFailureStore struct {
	store
	failure error
}

func (s *progressFailureStore) Progress(_ context.Context, _ projectApplicationPort.Scope, v *projectDomainEntity.Project) error {
	v.IssueCount, v.CompletedIssueCount = 7, 2
	return s.failure
}

func TestSourceHydrationFailurePreservesCanonicalMetadata(t *testing.T) {
	ctx := context.Background()
	gid, rid, pid := "C"+strings.Repeat("1", 32), strings.Repeat("3", 32), strings.Repeat("2", 32)
	hydration := errors.Join(projectDomainEntity.ErrSourceUnavailable, projectDomainEntity.ErrUnavailable)
	for _, operation := range []string{"get", "list", "update"} {
		t.Run(operation, func(t *testing.T) {
			st := &progressFailureStore{store: store{item: projectDomainEntity.Project{ID: pid, OwnerID: gid, DefaultRepositoryID: rid, CreatorID: "owner", Title: "owned metadata", Version: 1}}, failure: hydration}
			svc := &application.Service{Groups: &groups{}, Sources: &sources{space: gid}, Store: st}
			var got projectDomainEntity.Project
			var err error
			switch operation {
			case "get":
				got, err = svc.Get(ctx, "owner", gid, pid)
			case "list":
				var page projectApplicationDto.Page
				page, err = svc.List(ctx, "owner", gid, "")
				if len(page.Items) != 1 {
					t.Fatalf("own Project disappeared: %+v %v", page, err)
				}
				got = page.Items[0]
			case "update":
				got, err = svc.Update(ctx, "owner", gid, pid, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: "owned metadata"}, ExpectedVersion: 1})
			}
			if err != nil || got.ID != pid || got.Title != "owned metadata" || got.SourceAvailable || got.IssueCount != 0 || got.CompletedIssueCount != 0 {
				t.Fatalf("source hydration must not hide metadata or expose partial counts: %+v %v", got, err)
			}
		})
	}
	st := &progressFailureStore{store: store{item: projectDomainEntity.Project{ID: pid, OwnerID: gid, DefaultRepositoryID: rid, CreatorID: "owner", Title: "before", Version: 1}}, failure: hydration}
	svc := &application.Service{Groups: &groups{}, Sources: &sources{space: gid}, Store: st}
	updated, err := svc.Update(ctx, "owner", gid, pid, projectDomainEntity.Update{Content: projectDomainEntity.Content{Title: "after"}, ExpectedVersion: 1})
	if err != nil || updated.Version != 2 || st.item.Title != "after" || updated.SourceAvailable {
		t.Fatalf("committed own metadata update reported failure: %+v %v", updated, err)
	}
}

func TestOwnStorageCancellationAndLegacyHydrationStillFail(t *testing.T) {
	gid, rid, pid := "C"+strings.Repeat("1", 32), strings.Repeat("3", 32), strings.Repeat("2", 32)
	for _, failure := range []error{projectDomainEntity.ErrUnavailable, projectDomainEntity.ErrNotFound, context.Canceled, context.DeadlineExceeded, errors.Join(projectDomainEntity.ErrSourceUnavailable, context.Canceled), errors.Join(projectDomainEntity.ErrSourceUnavailable, context.DeadlineExceeded)} {
		st := &progressFailureStore{store: store{item: projectDomainEntity.Project{ID: pid, OwnerID: gid, DefaultRepositoryID: rid, Title: "own", Version: 1}}, failure: failure}
		svc := &application.Service{Groups: &groups{}, Sources: &sources{space: gid}, Store: st}
		if _, err := svc.Get(context.Background(), "owner", gid, pid); !errors.Is(err, failure) {
			t.Fatalf("own storage or cancellation error swallowed: %v => %v", failure, err)
		}
	}
	failure := errors.Join(projectDomainEntity.ErrSourceUnavailable, projectDomainEntity.ErrUnavailable)
	st := &progressFailureStore{store: store{item: projectDomainEntity.Project{ID: pid, OwnerID: gid, DefaultRepositoryID: rid, Title: "legacy", Version: 1}}, failure: failure}
	svc := &application.Service{Groups: &groups{}, Sources: &sources{space: gid}, Store: st}
	if _, err := svc.Get(context.Background(), "owner", rid, pid); !errors.Is(err, failure) {
		t.Fatalf("legacy source failure swallowed: %v", err)
	}
}

func (s *store) ListReferences(context.Context, projectApplicationPort.Scope, string, string, string, bool) (projectApplicationDto.ReferencePage, error) {
	return projectApplicationDto.ReferencePage{}, nil
}

func (s *store) ChangeReference(context.Context, projectApplicationPort.Scope, string, string, projectDomainEntity.Reference, bool, int64) (projectDomainEntity.Project, error) {
	return projectDomainEntity.Project{}, nil
}

func TestLegacyProjectScopeUsesGroupIDWithoutChangingPublicGroupID(t *testing.T) {
	gid := "C" + strings.Repeat("1", 32)
	rid := strings.Repeat("3", 32)
	for _, route := range []string{gid, rid} {
		t.Run(route, func(t *testing.T) {
			st := &store{}
			svc := &application.Service{Groups: &groups{}, Sources: &sources{space: gid}, Store: st}
			got, err := svc.Create(context.Background(), "owner", route, "group-context-key-0001", projectDomainEntity.Content{Title: "project", SourceRepositoryID: rid})
			if err != nil || got.OwnerID != gid || got.DefaultRepositoryID != rid {
				t.Fatalf("legacy public GroupID must remain GroupID, got %+v, %v", got, err)
			}
		})
	}
}

func (s *store) Inspections(context.Context, projectApplicationPort.Scope, string, string, string) ([]projectDomainEntity.InspectionSubmission, error) {
	return nil, nil
}

func TestPersonalProjectOwner(t *testing.T) {
	owner := "U" + strings.Repeat("a", 32)
	other := "U" + strings.Repeat("b", 32)
	svc := &application.Service{Groups: &groups{denied: true}, Sources: &sources{space: owner}, Store: &store{}}
	ctx := context.Background()
	project, err := svc.Create(ctx, owner, owner, "personal-project-0001", projectDomainEntity.Content{Title: "個人專案", SourceRepositoryID: strings.Repeat("c", 32)})
	if err != nil || project.OwnerID != owner {
		t.Fatalf("create personal project: %+v %v", project, err)
	}
	if _, err = svc.Get(ctx, other, owner, project.ID); !errors.Is(err, projectDomainEntity.ErrNotFound) {
		t.Fatalf("foreign personal owner: %v", err)
	}
}
