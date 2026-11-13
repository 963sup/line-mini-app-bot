package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/project"
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/line"
	"github.com/963sup/line-mini-app-bot/internal/business/space"
	spacedb "github.com/963sup/line-mini-app-bot/internal/business/space/firestore"
)

func newSpaces(ctx context.Context) (*space.Service, func() error, error) {
	cloud, channel, token := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if cloud == "" || channel == "" || token == "" {
		return nil, func() error { return nil }, nil
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, e := firestore.NewClientWithDatabase(ctx, cloud, database)
	if e != nil {
		return nil, nil, fmt.Errorf("cannot initialize space database")
	}
	return &space.Service{Store: &spacedb.Store{Client: client, ChannelID: channel}, Membership: &repositoryline.Membership{Token: token, Client: &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}}, client.Close, nil
}

type projectSources struct{ repositories *repository.Service }

func (s projectSources) Authorize(ctx context.Context, actor, id, expectedSpaceID string) (project.Source, error) {
	if s.repositories == nil {
		return project.Source{}, project.ErrUnavailable
	}
	r, e := s.repositories.Get(ctx, actor, id)
	if errors.Is(e, repository.ErrNotFound) {
		return project.Source{}, project.ErrNotFound
	}
	if e != nil {
		return project.Source{}, e
	}
	if expectedSpaceID != "" && r.GroupID != expectedSpaceID {
		return project.Source{}, project.ErrNotFound
	}
	return project.Source{ID: r.ID, SpaceID: r.GroupID}, nil
}

func newProjects(spaces *space.Service, repositories *repository.Service) *project.Service {
	if spaces == nil {
		return nil
	}
	store, ok := spaces.Store.(*spacedb.Store)
	if !ok {
		return nil
	}
	return &project.Service{Spaces: spaces, Sources: projectSources{repositories}, Store: &projectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}
