package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/discussion"
	discussiondb "github.com/963sup/line-mini-app-bot/internal/business/discussion/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/issue"
	issuedb "github.com/963sup/line-mini-app-bot/internal/business/issue/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	repositorydb "github.com/963sup/line-mini-app-bot/internal/business/repository/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/line"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/milestones"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/statistics"
)

func newIssues(repositories *repository.Service) *issue.Service {
	if repositories == nil {
		return nil
	}
	store, ok := repositories.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &issue.Service{Repositories: repositories, Store: &issuedb.Store{Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store, Milestones: store}}
}

func newLabels(repositories *repository.Service) *labels.Service {
	if repositories == nil {
		return nil
	}
	store, ok := repositories.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &labels.Service{Repositories: repositories, Store: store}
}

func newDiscussions(repositories *repository.Service) *discussion.Service {
	if repositories == nil {
		return nil
	}
	store, ok := repositories.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &discussion.Service{Repositories: repositories, Store: &discussiondb.Store{Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store}}
}

func newRepositories(ctx context.Context) (*repository.Service, func() error, error) {
	project, channel, token := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if project == "" || channel == "" || token == "" {
		return nil, func() error { return nil }, nil
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot initialize repository database")
	}
	return &repository.Service{Store: &repositorydb.Store{Client: client, ChannelID: channel}, Membership: &repositoryline.Membership{Token: token, Client: &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}}, client.Close, nil
}

func newMilestones(r *repository.Service) *milestones.Service {
	if r == nil {
		return nil
	}
	store, ok := r.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &milestones.Service{Repositories: r, Store: store}
}

func newStatistics(r *repository.Service) *statistics.Service {
	if r == nil {
		return nil
	}
	store, ok := r.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &statistics.Service{Repositories: r, Store: store}
}
