package application

import (
	"context"
	"errors"
	"regexp"

	groupentity "github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
	projectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/project/application/dto"
	projectApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/project/application/port"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/identity"
)

var creationKey = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type Service struct {
	Groups  projectApplicationPort.Groups
	Sources projectApplicationPort.Sources
	Store   projectApplicationPort.Store
}

func (s *Service) resolveProjectScope(ctx context.Context, actor, id string) (projectApplicationPort.Scope, error) {
	if s == nil || s.Groups == nil || s.Sources == nil || s.Store == nil {
		return projectApplicationPort.Scope{}, projectDomainEntity.ErrUnavailable
	}
	scope := projectApplicationPort.Scope{OwnerID: id, Actor: actor}
	if !identity.ValidOwner(id) {
		// Repository-scoped routes resolve and authorize their source first.
		if !projectDomainEntity.ValidID(id) {
			return projectApplicationPort.Scope{}, projectDomainEntity.ErrNotFound
		}
		source, err := s.Sources.Authorize(ctx, actor, id, "")
		if err != nil {
			return projectApplicationPort.Scope{}, err
		}
		scope.OwnerID, scope.SourceRepositoryID, scope.Legacy = source.OwnerID, source.ID, true
	}
	if identity.IsUserOwnerID(scope.OwnerID) {
		if actor != scope.OwnerID {
			return projectApplicationPort.Scope{}, projectDomainEntity.ErrNotFound
		}
		return scope, nil
	}
	value, err := s.Groups.Get(ctx, actor, scope.OwnerID)
	if errors.Is(err, groupentity.ErrNotFound) {
		return projectApplicationPort.Scope{}, projectDomainEntity.ErrNotFound
	}
	if err != nil {
		return projectApplicationPort.Scope{}, err
	}
	scope.OwnerID = value.ID
	return scope, nil
}

func (s *Service) progress(ctx context.Context, r projectApplicationPort.Scope, p *projectDomainEntity.Project) error {
	_, err := s.Sources.Authorize(ctx, r.Actor, p.DefaultRepositoryID, r.OwnerID)
	if err != nil {
		if r.Legacy || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		p.SourceAvailable = false
		p.IssueCount, p.CompletedIssueCount = 0, 0
		return nil
	}
	p.SourceAvailable = true
	err = s.Store.Progress(ctx, r, p)
	if !r.Legacy && errors.Is(err, projectDomainEntity.ErrSourceUnavailable) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		p.SourceAvailable = false
		p.IssueCount, p.CompletedIssueCount = 0, 0
		return nil
	}
	return err
}

func (s *Service) List(ctx context.Context, actor, rid, cursor string) (projectApplicationDto.Page, error) {
	r, e := s.resolveProjectScope(ctx, actor, rid)
	if e != nil {
		return projectApplicationDto.Page{}, e
	}
	if len(cursor) > 1024 {
		return projectApplicationDto.Page{}, projectDomainEntity.ErrInput
	}
	page, err := s.Store.List(ctx, r, actor, cursor)
	if err != nil {
		return projectApplicationDto.Page{}, err
	}
	for i := range page.Items {
		if err = s.progress(ctx, r, &page.Items[i]); err != nil {
			return projectApplicationDto.Page{}, err
		}
	}
	return page, nil
}

func (s *Service) Get(ctx context.Context, actor, rid, id string) (projectDomainEntity.Project, error) {
	r, e := s.resolveProjectScope(ctx, actor, rid)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	if !projectDomainEntity.ValidID(id) {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrNotFound
	}
	result, err := s.Store.Get(ctx, r, id)
	if err == nil {
		err = s.progress(ctx, r, &result)
	}
	return result, err
}

func (s *Service) Create(ctx context.Context, actor, rid, key string, c projectDomainEntity.Content) (projectDomainEntity.Project, error) {
	r, e := s.resolveProjectScope(ctx, actor, rid)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	if r.Legacy {
		c.SourceRepositoryID = r.SourceRepositoryID
	}
	if !projectDomainEntity.ValidID(c.SourceRepositoryID) {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	if _, e = s.Sources.Authorize(ctx, actor, c.SourceRepositoryID, r.OwnerID); e != nil {
		return projectDomainEntity.Project{}, e
	}
	r.SourceRepositoryID = c.SourceRepositoryID
	c, e = projectDomainEntity.Normalize(c)
	if e != nil || !creationKey.MatchString(key) {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	value, e := s.Store.Create(ctx, r, actor, key, c)
	if e == nil {
		e = s.progress(ctx, r, &value)
	}
	return value, e
}

func (s *Service) Update(ctx context.Context, actor, rid, id string, input projectDomainEntity.Update) (projectDomainEntity.Project, error) {
	r, e := s.resolveProjectScope(ctx, actor, rid)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	if !projectDomainEntity.ValidID(id) {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrNotFound
	}
	if input.ItemID != "" || input.ItemStatus != "" {
		if r.Legacy || !projectDomainEntity.ValidID(input.ItemID) || (input.ItemStatus != "todo" && input.ItemStatus != "in_progress" && input.ItemStatus != "done") || input.LinkRepositoryID != "" || input.UnlinkRepositoryID != "" {
			return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
		}
	}
	if input.LinkRepositoryID != "" || input.UnlinkRepositoryID != "" {
		if r.Legacy || (input.LinkRepositoryID != "" && input.UnlinkRepositoryID != "") {
			return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
		}
		if input.LinkRepositoryID != "" {
			if !projectDomainEntity.ValidID(input.LinkRepositoryID) {
				return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
			}
			r.SelectedSource, e = s.Sources.Authorize(ctx, actor, input.LinkRepositoryID, "")
			if e != nil {
				return projectDomainEntity.Project{}, e
			}
		} else if !projectDomainEntity.ValidID(input.UnlinkRepositoryID) {
			return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
		}
	}
	value, e := s.Store.RunUpdate(ctx, r, id, func(work projectApplicationPort.UpdateWork) error {
		return updateProject(work, actor, input)
	})
	if e == nil {
		e = s.progress(ctx, r, &value)
	}
	return value, e
}

func updateProject(work projectApplicationPort.UpdateWork, actor string, input projectDomainEntity.Update) error {
	current, err := work.Current()
	if err != nil {
		return err
	}
	updated, err := projectDomainEntity.ApplyUpdate(current, actor, input)
	if err != nil {
		return err
	}
	var writes projectApplicationPort.UpdateWrites
	if input.ItemID != "" {
		status, err := work.ItemStatus(input.ItemID)
		if err != nil {
			return err
		}
		if status != input.ItemStatus {
			if current.Version >= projectDomainEntity.MaxVersion {
				return projectDomainEntity.ErrConflict
			}
			writes.ItemID, writes.ItemStatus = input.ItemID, input.ItemStatus
			updated.Version = current.Version + 1
		}
	}
	linkID, add := input.LinkRepositoryID, true
	if linkID == "" {
		linkID, add = input.UnlinkRepositoryID, false
	}
	if linkID != "" {
		exists, err := work.RepositoryLinked(linkID)
		if err != nil {
			return err
		}
		if exists != add {
			if current.Version >= projectDomainEntity.MaxVersion {
				return projectDomainEntity.ErrConflict
			}
			if add {
				if err := work.CheckLinkSource(linkID); err != nil {
					return err
				}
			}
			writes.RepositoryID, writes.LinkRepository = linkID, add
			updated.Version = current.Version + 1
		}
	}
	return work.Save(actor, current, updated, writes)
}

func (s *Service) ListItems(ctx context.Context, actor, rid, id, cursor string) (projectDomainEntity.ItemPage, error) {
	r, e := s.resolveProjectScope(ctx, actor, rid)
	if e != nil {
		return projectDomainEntity.ItemPage{}, e
	}
	if !projectDomainEntity.ValidID(id) || len(cursor) > 1024 {
		return projectDomainEntity.ItemPage{}, projectDomainEntity.ErrInput
	}
	value, e := s.Store.Get(ctx, r, id)
	if e != nil {
		return projectDomainEntity.ItemPage{}, e
	}
	if _, e = s.Sources.Authorize(ctx, actor, value.DefaultRepositoryID, r.OwnerID); e != nil {
		return projectDomainEntity.ItemPage{}, e
	}
	r.SourceRepositoryID = value.DefaultRepositoryID
	return s.Store.ListItems(ctx, r, actor, id, cursor)
}

func (s *Service) ChangeItem(ctx context.Context, actor, rid, id, target string, add bool, version int64) (projectDomainEntity.Project, error) {
	r, e := s.resolveProjectScope(ctx, actor, rid)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	if !projectDomainEntity.ValidID(id) || !projectDomainEntity.ValidID(target) || version < 1 {
		return projectDomainEntity.Project{}, projectDomainEntity.ErrInput
	}
	value, e := s.Store.Get(ctx, r, id)
	if e != nil {
		return projectDomainEntity.Project{}, e
	}
	if _, e = s.Sources.Authorize(ctx, actor, value.DefaultRepositoryID, r.OwnerID); e != nil {
		return projectDomainEntity.Project{}, e
	}
	r.SourceRepositoryID = value.DefaultRepositoryID
	result, e := s.Store.ChangeItem(ctx, r, actor, id, target, add, version)
	if e == nil {
		e = s.progress(ctx, r, &result)
	}
	return result, e
}
