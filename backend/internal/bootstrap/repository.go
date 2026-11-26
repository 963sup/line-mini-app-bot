package bootstrap

import (
	"fmt"
	"net/http"
	"os"
	"time"

	operations "github.com/963sup/line-mini-app-bot/internal/business/object/application/document"

	httptransport "github.com/963sup/line-mini-app-bot/internal/transport/http"

	"cloud.google.com/go/firestore"
	discussiondb "github.com/963sup/line-mini-app-bot/internal/business/discussion/adapter/firestore"
	discussionapp "github.com/963sup/line-mini-app-bot/internal/business/discussion/application"
	issuedb "github.com/963sup/line-mini-app-bot/internal/business/issue/adapter/firestore"
	issueapplication "github.com/963sup/line-mini-app-bot/internal/business/issue/application"
	objectdb "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/firestore"
	candidateai "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/openai"
	"github.com/963sup/line-mini-app-bot/internal/business/object/adapter/poppler"
	objectApplication "github.com/963sup/line-mini-app-bot/internal/business/object/application"
	organizationdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	repositorydb "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore"
	businessspacedb "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore/space"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationLabels "github.com/963sup/line-mini-app-bot/internal/business/repository/application/labels"
	repositoryApplicationMilestones "github.com/963sup/line-mini-app-bot/internal/business/repository/application/milestones"
	repositoryApplicationSpace "github.com/963sup/line-mini-app-bot/internal/business/repository/application/space"
	repositoryApplicationStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/application/statistics"
	repositoryApplicationWiki "github.com/963sup/line-mini-app-bot/internal/business/repository/application/wiki"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func newIssues(repositories *repositoryApplication.Service, store *repositorydb.Store) *issueapplication.Service {
	if store == nil {
		return nil
	}
	return issueapplication.New(repositories, &issuedb.Store{RepositoryWritable: store.Writable, Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store, Milestones: store})
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
	return discussionapp.New(repositories, &discussiondb.Store{RepositoryWritable: store.Writable, Client: store.Client, ChannelID: store.ChannelID, Numbers: store, Labels: store})
}

func newRepositories(firestoreConfigured bool, getFirestore func() (*firestore.Client, error), channel string, ownerAccess *organizationdb.Store) (*repositoryApplication.Service, *repositorydb.Store, error) {
	if !firestoreConfigured || channel == "" || ownerAccess == nil {
		return nil, nil, nil
	}
	client, err := getFirestore()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot initialize repository database")
	}
	projects := &projectdb.Store{Client: client, ChannelID: channel}
	store := &repositorydb.Store{OrganizationVisible: ownerAccess.VisibleTo, Client: client, ChannelID: channel, CountIssues: issuedb.CountRepositoryIssues, CountDiscussions: discussiondb.CountRepositoryDiscussions, CountProjects: projects.CountRepositoryProjects}
	return &repositoryApplication.Service{Store: store, AddressWriter: store, OwnerAccess: ownerAccess}, store, nil
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
	objects := &objectdb.Store{RepositoryWritable: store.Writable, QueueQuantityEvents: os.Getenv("OBJECT_QUANTITY_PUBSUB_TOPIC") != "", Client: store.Client, ChannelID: store.ChannelID}
	service := &objectApplication.Service{QuantityHistory: objects, Templates: objects, PersonalTemplates: objects, Repositories: r, Store: objects}
	if key, model := os.Getenv("OPENAI_API_KEY"), os.Getenv("AI_DOCUMENT_MODEL"); key != "" && model != "" {
		client := openai.NewClient(option.WithAPIKey(key), option.WithBaseURL("https://api.openai.com/v1/"), option.WithMaxRetries(0), option.WithHTTPClient(&http.Client{Transport: httptransport.TraceTransport{}, Timeout: 2 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}))
		extractor := &candidateai.Extractor{Responses: client.Responses, Model: model}
		bridge := &operations.Preparation{Extract: extractor, Text: &poppler.Extractor{}, Procedures: extractor}
		service.CandidatePreparer = bridge
		service.TemplatePreparer = bridge
	}
	return service
}

func newRepositorySpaces(r *repositoryApplication.Service, store *repositorydb.Store) *repositoryApplicationSpace.Service {
	if store == nil {
		return nil
	}
	return &repositoryApplicationSpace.Service{Repositories: r, Store: &businessspacedb.Store{RepositoryWritable: store.Writable, Client: store.Client, ChannelID: store.ChannelID}}
}

func newWiki(repositories *repositoryApplication.Service, store *repositorydb.Store) *repositoryApplicationWiki.Service {
	if store == nil {
		return nil
	}
	return &repositoryApplicationWiki.Service{Repositories: repositories, Store: store}
}
