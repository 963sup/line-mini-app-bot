package milestones

import (
	repositoryDomainEntityMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/milestones"
)

type Page struct {
	Items      []repositoryDomainEntityMilestones.Milestone
	NextCursor string
}
