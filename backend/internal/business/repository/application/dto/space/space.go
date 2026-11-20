package space

import (
	repositoryDomainEntitySpace "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/space"
)

type Page struct {
	Items      []repositoryDomainEntitySpace.Space
	NextCursor string
}
