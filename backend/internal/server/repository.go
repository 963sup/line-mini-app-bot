package server

import (
	objectimportApplication "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application"
	objectimportApplicationAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/automatic"
	objectimportApplicationParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/parsing"
	objectimportApplicationPortAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/automatic"
	objectimportApplicationPortParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/parsing"
)

import (
	objectApplication "github.com/963sup/line-mini-app-bot/internal/business/object/application"
)

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/repository/wiki"

	"cloud.google.com/go/firestore"
	discussiondb "github.com/963sup/line-mini-app-bot/internal/business/discussion/adapter/firestore"
	discussionapp "github.com/963sup/line-mini-app-bot/internal/business/discussion/application"
	issuedb "github.com/963sup/line-mini-app-bot/internal/business/issue/adapter/firestore"
	issueapplication "github.com/963sup/line-mini-app-bot/internal/business/issue/application"

	objectdb "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/firestore"

	importdb "github.com/963sup/line-mini-app-bot/internal/business/objectimport/adapter/firestore"

	parseropenai "github.com/963sup/line-mini-app-bot/internal/business/objectimport/adapter/openai/parsing"
	"github.com/963sup/line-mini-app-bot/internal/business/repository"
	repositorydb "github.com/963sup/line-mini-app-bot/internal/business/repository/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/labels"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/milestones"
	businessspace "github.com/963sup/line-mini-app-bot/internal/business/repository/space"
	businessspacedb "github.com/963sup/line-mini-app-bot/internal/business/repository/space/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/statistics"
	"github.com/963sup/line-mini-app-bot/internal/business/subobject"
	subobjectdb "github.com/963sup/line-mini-app-bot/internal/business/subobject/firestore"
	linetransport "github.com/963sup/line-mini-app-bot/internal/transport/line"
)

func newIssues(repositories *repository.Service) *issueapplication.Service {
	if repositories == nil {
		return nil
	}
	store, ok := repositories.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return issueapplication.New(repositories, &issuedb.Store{Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store, Milestones: store})
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

func newDiscussions(repositories *repository.Service) *discussionapp.Service {
	if repositories == nil {
		return nil
	}
	store, ok := repositories.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return discussionapp.New(repositories, &discussiondb.Store{Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store})
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
	return &repository.Service{Store: &repositorydb.Store{Client: client, ChannelID: channel}, Membership: &linetransport.Membership{Token: token, Client: &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}}, client.Close, nil
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

func newObjects(r *repository.Service) *objectApplication.Service {
	if r == nil {
		return nil
	}
	store, ok := r.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &objectApplication.Service{Repositories: r, Store: &objectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newObjectImports(r *repository.Service) *objectimportApplication.Service {
	if r == nil {
		return nil
	}
	store, ok := r.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &objectimportApplication.Service{Repositories: r, Objects: newObjects(r), Store: &importdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newObjectParser(r *repository.Service) *objectimportApplicationParsing.Service {
	if r == nil {
		return nil
	}
	var producer objectimportApplicationPortParsing.Producer
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		producer = &parseropenai.Producer{Key: key, Model: os.Getenv("OBJECT_IMPORT_MODEL")}
	}
	return &objectimportApplicationParsing.Service{Repositories: r, Producer: producer}
}

func newAutomaticObjectImports(r *repository.Service) *objectimportApplicationAutomatic.Service {
	if r == nil {
		return nil
	}
	store, ok := r.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	var producer objectimportApplicationPortAutomatic.Producer
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		producer = &parseropenai.Producer{Key: key, Model: os.Getenv("OBJECT_IMPORT_MODEL")}
	}
	return &objectimportApplicationAutomatic.Service{Repositories: r, Objects: newObjects(r), Store: &importdb.AutomaticStore{Store: importdb.Store{Client: store.Client, ChannelID: store.ChannelID}}, Producer: producer}
}

func newSubobjects(r *repository.Service) *subobject.Service {
	if r == nil {
		return nil
	}
	store, ok := r.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &subobject.Service{Repositories: r, Store: &subobjectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newRepositorySpaces(r *repository.Service) *businessspace.Service {
	if r == nil {
		return nil
	}
	store, ok := r.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &businessspace.Service{Repositories: r, Store: &businessspacedb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newWiki(repositories *repository.Service) *wiki.Service {
	if repositories == nil {
		return nil
	}
	store, ok := repositories.Store.(*repositorydb.Store)
	if !ok {
		return nil
	}
	return &wiki.Service{Repositories: repositories, Store: store}
}
