package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/objectimport/adapter/workbook"
	workflow "github.com/963sup/line-mini-app-bot/internal/business/workflow/application"

	"cloud.google.com/go/firestore"
	discussiondb "github.com/963sup/line-mini-app-bot/internal/business/discussion/adapter/firestore"
	discussionapp "github.com/963sup/line-mini-app-bot/internal/business/discussion/application"
	issuedb "github.com/963sup/line-mini-app-bot/internal/business/issue/adapter/firestore"
	issueapplication "github.com/963sup/line-mini-app-bot/internal/business/issue/application"
	objectdb "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/firestore"
	objectApplication "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	importdb "github.com/963sup/line-mini-app-bot/internal/business/objectimport/adapter/firestore"
	parseropenai "github.com/963sup/line-mini-app-bot/internal/business/objectimport/adapter/openai/parsing"
	objectimportApplication "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application"
	objectimportApplicationAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/automatic"
	objectimportApplicationParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/parsing"
	objectimportApplicationPortAutomatic "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/automatic"
	objectimportApplicationPortParsing "github.com/963sup/line-mini-app-bot/internal/business/objectimport/application/port/parsing"
	repositorydb "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore"
	businessspacedb "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore/space"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/application/labels"
	repositoryApplicationMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/application/milestones"
	repositoryApplicationSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/space"
	repositoryApplicationStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/application/statistics"
	repositoryApplicationWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/application/wiki"
	subobjectdb "github.com/963sup/line-mini-app-bot/internal/business/subobject/adapter/firestore"
	subobjectApplication "github.com/963sup/line-mini-app-bot/internal/business/subobject/application"
	linetransport "github.com/963sup/line-mini-app-bot/internal/transport/line"
)

func newIssues(repositories *repositoryApplication.Service, store *repositorydb.Store) *issueapplication.Service {
	if store == nil {
		return nil
	}
	return issueapplication.New(repositories, &issuedb.Store{Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store, Milestones: store})
}

func newLabels(repositories *repositoryApplication.Service, store *repositorydb.Store) *repositoryApplicationLabels.Service {
	if store == nil {
		return nil
	}
	return &repositoryApplicationLabels.Service{Repositories: repositories, Store: store}
}

func newDiscussions(repositories *repositoryApplication.Service, store *repositorydb.Store) *discussionapp.Service {
	if store == nil {
		return nil
	}
	return discussionapp.New(repositories, &discussiondb.Store{Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store})
}

func newRepositories(ctx context.Context) (*repositoryApplication.Service, *repositorydb.Store, func() error, error) {
	project, channel, token := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN")
	if project == "" || channel == "" || token == "" {
		return nil, nil, func() error { return nil }, nil
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot initialize repository database")
	}
	store := &repositorydb.Store{Client: client, ChannelID: channel}
	return &repositoryApplication.Service{Store: store, Membership: &linetransport.Membership{Token: token, Client: &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}}, store, client.Close, nil
}

func newMilestones(r *repositoryApplication.Service, store *repositorydb.Store) *repositoryApplicationMilestones.Service {
	if store == nil {
		return nil
	}
	return &repositoryApplicationMilestones.Service{Repositories: r, Store: store}
}

func newStatistics(r *repositoryApplication.Service, store *repositorydb.Store) *repositoryApplicationStatistics.Service {
	if store == nil {
		return nil
	}
	return &repositoryApplicationStatistics.Service{Repositories: r, Store: store}
}

func newObjects(r *repositoryApplication.Service, store *repositorydb.Store) *objectApplication.Service {
	if store == nil {
		return nil
	}
	return &objectApplication.Service{Repositories: r, Store: &objectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newObjectImports(r *repositoryApplication.Service, store *repositorydb.Store) *objectimportApplication.Service {
	if store == nil {
		return nil
	}
	return &objectimportApplication.Service{Repositories: r, Objects: newObjects(r, store), Store: &importdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newObjectParser(r *repositoryApplication.Service) workflow.PDF {
	return workflow.PDF{Extract: newCandidateService(r).Parse}
}

func newCandidateService(r *repositoryApplication.Service) *objectimportApplicationParsing.Service {
	if r == nil {
		return nil
	}
	var producer objectimportApplicationPortParsing.Producer
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		producer = &parseropenai.Producer{Key: key, Model: os.Getenv("OBJECT_IMPORT_MODEL")}
	}
	return &objectimportApplicationParsing.Service{Repositories: r, Producer: producer}
}

func newAutomaticObjectImports(r *repositoryApplication.Service, store *repositorydb.Store) *objectimportApplicationAutomatic.Service {
	if store == nil {
		return nil
	}
	var producer objectimportApplicationPortAutomatic.Producer
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		producer = &parseropenai.Producer{Key: key, Model: os.Getenv("OBJECT_IMPORT_MODEL")}
	}
	return &objectimportApplicationAutomatic.Service{Workbook: workbook.Read, TextTable: workbook.ReadText, Repositories: r, Objects: newObjects(r, store), Store: &importdb.AutomaticStore{Store: importdb.Store{Client: store.Client, ChannelID: store.ChannelID}}, Producer: producer}
}

func newSubobjects(r *repositoryApplication.Service, store *repositorydb.Store) *subobjectApplication.Service {
	if store == nil {
		return nil
	}
	return &subobjectApplication.Service{Repositories: r, Store: &subobjectdb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newRepositorySpaces(r *repositoryApplication.Service, store *repositorydb.Store) *repositoryApplicationSpace.Service {
	if store == nil {
		return nil
	}
	return &repositoryApplicationSpace.Service{Repositories: r, Store: &businessspacedb.Store{Client: store.Client, ChannelID: store.ChannelID}}
}

func newWiki(repositories *repositoryApplication.Service, store *repositorydb.Store) *repositoryApplicationWiki.Service {
	if store == nil {
		return nil
	}
	return &repositoryApplicationWiki.Service{Repositories: repositories, Store: store}
}
