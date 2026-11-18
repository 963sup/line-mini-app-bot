package project

import (
	"context"
	"errors"
)

// Reference retains source identity only; display fields are resolved for each read.
type Reference struct {
	ID, SourceType, SourceID, ParentObjectID, SourceRepositoryID      string
	SourceUnavailable                                                 bool
	SourceTitle, SourceState, Quantity, Unit, CompletedQuantity, Type string
}
type ReferencePage struct {
	Items      []Reference
	NextCursor string
}

func ValidReference(r Reference) bool {
	if !ValidID(r.SourceID) {
		return false
	}
	switch r.SourceType {
	case "subobject":
		return ValidID(r.ParentObjectID)
	case "object", "issue", "discussion":
		return r.ParentObjectID == ""
	default:
		return false
	}
}

func (s *Service) referenceScope(ctx context.Context, actor, groupContextID, pid string) (Scope, Project, error) {
	r, e := s.scope(ctx, actor, groupContextID)
	if e != nil {
		return Scope{}, Project{}, e
	}
	if r.Legacy || !ValidID(pid) {
		return Scope{}, Project{}, ErrInput
	}
	p, e := s.Store.Get(ctx, r, pid)
	if e != nil {
		return Scope{}, Project{}, e
	}
	r.SourceRepositoryID = p.SourceRepositoryID
	return r, p, nil
}

func (s *Service) ListReferences(ctx context.Context, actor, groupContextID, pid, cursor string) (ReferencePage, error) {
	r, _, e := s.referenceScope(ctx, actor, groupContextID, pid)
	if e != nil {
		return ReferencePage{}, e
	}
	if len(cursor) > 1024 {
		return ReferencePage{}, ErrInput
	}
	_, e = s.Sources.Authorize(ctx, actor, r.SourceRepositoryID, r.GroupID)
	if errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) {
		return ReferencePage{}, e
	}
	return s.Store.ListReferences(ctx, r, actor, pid, cursor, e == nil)
}

func (s *Service) AddReference(ctx context.Context, actor, groupContextID, pid string, ref Reference, version int64) (Project, error) {
	r, p, e := s.referenceScope(ctx, actor, groupContextID, pid)
	if e != nil {
		return Project{}, e
	}
	if p.CreatorID != actor {
		return Project{}, ErrForbidden
	}
	if !ValidReference(ref) || version < 1 {
		return Project{}, ErrInput
	}
	if _, e = s.Sources.Authorize(ctx, actor, r.SourceRepositoryID, r.GroupID); e != nil {
		return Project{}, e
	}
	result, e := s.Store.ChangeReference(ctx, r, actor, pid, ref, true, version)
	if e == nil {
		e = s.progress(ctx, r, &result)
	}
	return result, e
}

func (s *Service) RemoveReference(ctx context.Context, actor, groupContextID, pid, itemID string, version int64) (Project, error) {
	r, p, e := s.referenceScope(ctx, actor, groupContextID, pid)
	if e != nil {
		return Project{}, e
	}
	if p.CreatorID != actor {
		return Project{}, ErrForbidden
	}
	if !ValidID(itemID) || version < 1 {
		return Project{}, ErrInput
	}
	result, e := s.Store.ChangeReference(ctx, r, actor, pid, Reference{ID: itemID}, false, version)
	if e == nil {
		e = s.progress(ctx, r, &result)
	}
	return result, e
}
