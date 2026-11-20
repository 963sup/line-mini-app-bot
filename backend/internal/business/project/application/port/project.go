package port

import (
	"context"

	groupentity "github.com/963sup/line-mini-app-bot/internal/business/group/domain/entity"
	projectApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/project/application/dto"
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

type Groups interface {
	Get(context.Context, string, string) (groupentity.Group, error)
}

type Sources interface {
	Authorize(context.Context, string, string, string) (projectDomainEntity.Source, error)
}

type Store interface {
	List(context.Context, Scope, string, string) (projectApplicationDto.Page, error)
	Get(context.Context, Scope, string) (projectDomainEntity.Project, error)
	Create(context.Context, Scope, string, string, projectDomainEntity.Content) (projectDomainEntity.Project, error)
	Update(context.Context, Scope, string, string, projectDomainEntity.Update) (projectDomainEntity.Project, error)
	ListItems(context.Context, Scope, string, string, string) (projectDomainEntity.ItemPage, error)
	ChangeItem(context.Context, Scope, string, string, string, bool, int64) (projectDomainEntity.Project, error)
	Progress(context.Context, Scope, *projectDomainEntity.Project) error
	Inspections(context.Context, Scope, string, string, string) ([]projectDomainEntity.InspectionSubmission, error)
	ListReferences(context.Context, Scope, string, string, string, bool) (projectApplicationDto.ReferencePage, error)
	ChangeReference(context.Context, Scope, string, string, projectDomainEntity.Reference, bool, int64) (projectDomainEntity.Project, error)
}

type Scope struct {
	OwnerID, SourceRepositoryID, Actor string
	Legacy                             bool
	Sources                            Sources
	SelectedSource                     projectDomainEntity.Source
}

// ReferenceSource resolves a reference against the default or explicitly authorized source.
func (r Scope) ReferenceSource(repositoryID string) (projectDomainEntity.Source, error) {
	if repositoryID == r.SourceRepositoryID {
		return projectDomainEntity.Source{ID: repositoryID, OwnerID: r.OwnerID}, nil
	}
	if repositoryID != r.SelectedSource.ID {
		return projectDomainEntity.Source{}, projectDomainEntity.ErrForbidden
	}
	return r.SelectedSource, nil
}
