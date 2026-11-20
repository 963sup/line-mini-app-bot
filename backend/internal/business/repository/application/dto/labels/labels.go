package labels

import (
	repositoryDomainEntityLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/labels"
)

type Page struct {
	Items      []repositoryDomainEntityLabels.Label
	NextCursor string
}
