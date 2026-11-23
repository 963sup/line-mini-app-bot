package bootstrap

import (
	"cloud.google.com/go/firestore"
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	source "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/repository"
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
)

func newProjects(client *firestore.Client, channel string, repositories *repositoryApplication.Service) *projectApplication.Service {
	if client == nil || channel == "" {
		return nil
	}
	projects := &projectdb.Store{Client: client, ChannelID: channel}
	return &projectApplication.Service{AuthorizeOrganization: projects.AuthorizeOrganization, Sources: source.Sources{Repositories: repositories}, Store: projects}
}
