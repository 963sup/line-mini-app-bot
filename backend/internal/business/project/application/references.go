package application

import (
	"context"
	"errors"

	projectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/project/application/dto"
	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

func (s *Service) projectScope(ctx context.Context, actor, ownerID, pid string) (projectApplicationPort.Scope, projectDomainEntity.Project, error) {
	r, e := s.resolveProjectScope(ctx, actor, ownerID)
	if e != nil {
		return projectApplicationPort.Scope{}, projectDomainEntity.Project{}, e
	}
	if r.Legacy || !projectDomainEntity.ValidID(pid) {
		return projectApplicationPort.Scope{}, projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	p, e := s.Store.Get(ctx, r, pid)
	if e != nil {
		return projectApplicationPort.Scope{}, projectDomainEntity.Project{}, e
	}
	r.Sources = s.Sources
	r.SourceRepositoryID = p.DefaultRepositoryID
	return r, p, nil
}

func (s *Service) ListReferences(ctx context.Context, actor, ownerID, pid, cursor string) (projectApplicationDto.ReferencePage, error) {
	r, _, e := s.projectScope(ctx, actor, ownerID, pid)
	if e != nil {
		return projectApplicationDto.ReferencePage{}, e
	}
	if len(cursor) > 1024 {
		return projectApplicationDto.ReferencePage{}, projectDomainEntity.ErrInput
	}
	_, e = s.Sources.Authorize(ctx, actor, r.SourceRepositoryID, r.OwnerID)
	if errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) {
		return projectApplicationDto.ReferencePage{}, e
	}
	return s.Store.ListReferences(ctx, r, actor, pid, cursor, e == nil)
}

func (s *Service) AddReference(ctx context.Context, actor, ownerID, pid string, ref projectDomainEntity.Reference, version int64) (projectDomainEntity.Project, error) {
	r, p, e := s.projectScope(ctx, actor, ownerID, pid)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	if p.CreatorID != actor {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrForbidden
	}
	if !projectDomainEntity.ValidReference(ref) || version < 1 {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	sourceID := ref.SourceRepositoryID
	expectedOwner := ""
	if sourceID == "" || sourceID == r.SourceRepositoryID {
		sourceID = r.SourceRepositoryID
		expectedOwner = r.OwnerID
	} else if !projectDomainEntity.ValidID(sourceID) {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	r.SelectedSource, e = s.Sources.Authorize(ctx, actor, sourceID, expectedOwner)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	ref.SourceRepositoryID = sourceID
	result, e := s.Store.ChangeReference(ctx, r, actor, pid, ref, true, version)
	if e == nil {
		e = s.progress(ctx, r, &result)
	}
	return result, e
}

func (s *Service) RemoveReference(ctx context.Context, actor, ownerID, pid, itemID string, version int64) (projectDomainEntity.Project, error) {
	r, p, e := s.projectScope(ctx, actor, ownerID, pid)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	if p.CreatorID != actor {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrForbidden
	}
	if !projectDomainEntity.ValidID(itemID) || version < 1 {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	result, e := s.Store.ChangeReference(ctx, r, actor, pid, projectDomainEntity.Reference{ID: itemID}, false, version)
	if e == nil {
		e = s.progress(ctx, r, &result)
	}
	return result, e
}

// ReferenceState checks the current state, not the outcome of a past request.
func (s *Service) ReferenceState(ctx context.Context, actor, ownerID, pid string, ref projectDomainEntity.Reference, desired bool) (present, matches bool, err error) {
	r, p, err := s.projectScope(ctx, actor, ownerID, pid)
	if err != nil {
		return false, false, err
	}
	if p.CreatorID != actor {
		return false, false, projectDomainEntity.ErrForbidden
	}
	if ref.ID != "" {
		if !projectDomainEntity.ValidID(ref.ID) || ref.SourceRepositoryID != "" || ref.SourceType != "" || ref.SourceID != "" || ref.ParentObjectID != "" {
			return false, false, projectDomainEntity.ErrInput
		}
	} else if !projectDomainEntity.ValidReference(ref) || !projectDomainEntity.ValidID(ref.SourceRepositoryID) {
		return false, false, projectDomainEntity.ErrInput
	}
	store, ok := s.Store.(projectApplicationPort.ReferenceStateStore)
	if !ok {
		return false, false, projectDomainEntity.ErrUnavailable
	}
	present, err = store.ReferencePresent(ctx, r, actor, pid, ref)
	return present, err == nil && present == desired, err
}
