package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	attendancehttp "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/http"
	attendanceline "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/line"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/presentation"
	discussionhttp "github.com/963sup/line-mini-app-bot/internal/business/discussion/adapter/http"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"
	enterprisehttp "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/http"
	explorehttp "github.com/963sup/line-mini-app-bot/internal/business/explore/adapter/http"
	explorerepository "github.com/963sup/line-mini-app-bot/internal/business/explore/adapter/repository"
	explorequery "github.com/963sup/line-mini-app-bot/internal/business/explore/application/query"
	issuehttp "github.com/963sup/line-mini-app-bot/internal/business/issue/adapter/http"
	journalhttp "github.com/963sup/line-mini-app-bot/internal/business/journal/adapter/http"
	membershipdb "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/firestore"
	membershiphttp "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/http"
	membershipapp "github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	objectevents "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/firestore"
	objecthttp "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/http"
	operationshttp "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/http/document"
	objectmessages "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/messaging"
	poppler "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/poppler"
	objectcommand "github.com/963sup/line-mini-app-bot/internal/business/object/application/command"
	operations "github.com/963sup/line-mini-app-bot/internal/business/object/application/document"
	organizationdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	organizationhttp "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/http"
	organizationline "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/line"
	projecthttp "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/http"
	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repositoryhttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http"
	labelhttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/labels"
	milestonehttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/milestones"
	businessgrouphttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/space"
	statisticshttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/statistics"
	wikihttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/wiki"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/line"
	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
	teamdb "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore"
	teamhttp "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/http"
	grouphttp "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/http/group"
	groupline "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/line"
	firestoreplatform "github.com/963sup/line-mini-app-bot/internal/platform/firestore"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	httptransport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/browser"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/geolocation"
	"google.golang.org/api/pubsub/v1"
	"sync"
)

func Run() (resultErr error) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	shutdownTelemetry := httptransport.NewTelemetry(slog.Default())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resultErr = errors.Join(resultErr, shutdownTelemetry(ctx))
	}()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	geolocation.RegisterGeolocation(mux)
	if err := browser.Register(mux); err != nil {
		return err
	}
	auth, endpointURL, err := newMiniAppAuthenticator(os.Getenv("LINE_MINI_APP_ENV"))
	if err != nil {
		return err
	}
	// Open Firestore only when an enabled module needs it. Run owns the
	// client through startup and server shutdown.
	project, database := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("FIRESTORE_DATABASE_ID")
	firestoreConfigured := project != ""
	var db *firestore.Client
	getFirestore := func() (*firestore.Client, error) {
		if db != nil {
			return db, nil
		}
		client, err := firestoreplatform.NewClient(context.Background(), project, database)
		if err != nil {
			return nil, err
		}
		db = client
		return db, nil
	}
	defer func() {
		if db != nil {
			if closeErr := db.Close(); resultErr == nil {
				resultErr = closeErr
			}
		}
	}()
	sessions, err := newSessions(firestoreConfigured, getFirestore, auth.ChannelID)
	if err != nil {
		return err
	}
	auth.Sessions = sessions
	auth.Register(mux)
	appURL := lineplatform.MiniAppURL(auth.LIFFID)
	lineConfig := messagingEnvironment()
	if err := lineConfig.validate(); err != nil {
		return err
	}
	attendance, err := newAttendance(firestoreConfigured, getFirestore, lineConfig.channel)
	if err != nil {
		return err
	}
	var groupLookup func(context.Context, string) string
	lookup := func(ctx context.Context, group string) string {
		if groupLookup != nil {
			return groupLookup(ctx, group)
		}
		return ""
	}
	var punches *attendanceline.Attendance
	var browserAttendance *attendancehttp.Attendance
	if attendance != nil {
		punches = &attendanceline.Attendance{Store: attendance.store, Zone: attendance.zone, GroupName: lookup}
		punches.AppURL = appURL
		membership := lineMembership(lineConfig.token, 5*time.Second)
		browserAttendance = attendancehttp.NewAttendance(attendance.store, attendance.zone, membership, func(ctx context.Context, group string) string {
			return presentation.GroupName(group, lookup(ctx, group))
		})
	}

	attendancehttp.RegisterAttendanceAPI(mux, browserAttendance, auth)
	operationshttp.Register(mux, &operations.DocumentTextService{Extractor: &poppler.Extractor{}}, auth)
	var messaging *lineplatform.Messaging
	if token := lineConfig.token; token != "" {
		messaging, err = lineplatform.NewMessaging(token, &http.Client{Transport: httptransport.TraceTransport{}, Timeout: 10 * time.Second})
		if err != nil {
			return err
		}
	}
	summaries, recoverSummaries, err := registerSummary(context.Background(), mux, messaging, lineConfig.channel, firestoreConfigured, getFirestore)
	if err != nil {
		return err
	}
	membershipLookup := lineMembership(lineConfig.token, time.Second)
	var repositoryOwners repositoryApplicationPort.OwnerAccess
	var resolveRepositoryOwner func(context.Context, string) (string, string, error)
	if sessions != nil && lineConfig.channel != "" {
		ownerContext := organizationdb.Store{DB: db, Channel: lineConfig.channel}
		repositoryOwners = ownerContext
		resolveRepositoryOwner = ownerContext.ResolveRepositoryOwner
	}
	repositories, repositoryStore, err := newRepositories(firestoreConfigured, getFirestore, lineConfig.channel, repositoryOwners)
	if err != nil {
		return err
	}
	groups, groupStore, err := newGroups(firestoreConfigured, getFirestore, lineConfig.channel, membershipLookup)
	if err != nil {
		return err
	}
	var lineTeams *teamdb.LINEObservation
	if groupStore != nil {
		lineTeams = &teamdb.LINEObservation{Store: teamdb.Store{
			DB: groupStore.Client, Channel: groupStore.ChannelID,
		}, Membership: groups.Membership}
	}
	var creation organizationline.Conversation
	var invitations *membershipapp.Service
	organizationHandler := organizationhttp.Handler{Auth: auth}
	enterpriseHandler := enterprisehttp.Handler{Auth: auth}
	teamHandler := teamhttp.Handler{Auth: auth}
	if sessions != nil && lineConfig.channel != "" {
		channel := lineConfig.channel
		organizationStore := organizationdb.Store{
			DB: db, Channel: channel,
			EnterpriseNavigation: enterprisedb.NavigationFacts,
			CanCreateEnterprise:  enterprisedb.CanStartCreation,
			CreateEnterprise:     enterprisedb.CreateFromConversation,
		}
		organizationHandler.Store = organizationStore
		enterpriseStore := enterprisedb.Store{
			DB: db, Channel: channel,
			OrganizationMember: membershipdb.OrganizationMember,
			TeamMember:         membershipdb.TeamMember,
		}
		teamStore := teamdb.Store{
			DB: db, Channel: channel,
			OrganizationMember: membershipdb.OrganizationMember,
			TeamMember:         membershipdb.TeamMember,
			OpenMembership:     membershipdb.OpenTeamMembership,
		}
		if lineTeams != nil {
			enterpriseStore.Groups = lineTeams
			teamStore.Groups = lineTeams
			teamStore.LINEActivationMember = lineTeams.Membership
		}
		enterpriseHandler.Store = enterpriseStore
		teamHandler.Store = teamStore
		creation = organizationStore
		invitations = membershipInvitations(db, channel, lineConfig.secret)
	}
	httpapi.RegisterOrganizations(mux, organizationHandler)
	httpapi.RegisterMembership(mux, membershiphttp.Handler{Service: invitations, Auth: auth})
	httpapi.RegisterEnterprises(mux, enterpriseHandler)
	httpapi.RegisterTeams(mux, teamHandler)
	var groupCatalog grouphttp.Catalog
	if groupStore != nil {
		teamStore := teamdb.Store{
			DB: groupStore.Client, Channel: groupStore.ChannelID, TeamMember: membershipdb.TeamMember,
		}
		groupCatalog = teamdb.GroupCatalog{Store: teamStore}
	}
	var projects *projectApplication.Service
	if sessions != nil {
		projects = newProjects(db, lineConfig.channel, repositories)
	}
	journalhttp.Register(mux, newJournals(groupCatalog, groupStore, repositories, projects), auth)
	grouphttp.Register(mux, groupCatalog, auth)
	repositoryhttp.Register(mux, repositories, auth)
	explorehttp.Register(mux, explorequery.Service{Repositories: explorerepository.Source{Discovery: repositoryApplication.Discovery{Store: repositoryStore}}}, auth)
	wikihttp.Register(mux, newWiki(repositories, repositoryStore), auth)
	labelhttp.Register(mux, newLabels(repositories, repositoryStore), auth)
	milestonehttp.Register(mux, newMilestones(repositories, repositoryStore), auth)
	projecthttp.Register(mux, projects, auth)
	statisticshttp.Register(mux, newStatistics(repositories, repositoryStore), auth)
	issuehttp.Register(mux, newIssues(repositories, repositoryStore), auth)
	businessgrouphttp.Register(mux, newRepositorySpaces(repositories, repositoryStore), auth)
	objecthttp.Register(mux, newObjects(repositories, repositoryStore), auth)
	discussionhttp.Register(mux, newDiscussions(repositories, repositoryStore), auth)
	var repositoryObserver *repositoryline.Observer
	if repositories != nil {
		repositoryObserver = &repositoryline.Observer{Store: repositories.Store, ResolveOwner: resolveRepositoryOwner}
	}
	var groupObserver *groupline.Observer
	if lineTeams != nil {
		groupObserver = &groupline.Observer{Store: lineTeams}
	}
	groupLookup = registerLINE(mux, lineConfig, messaging, appURL, endpointURL, punches, summaries, repositoryObserver, groupObserver, repositories, creation, invitations)
	server := &http.Server{
		Addr: net.JoinHostPort("", port), Handler: httptransport.RequestDiagnostics(mux),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 100 * time.Second, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	var workers sync.WaitGroup
	defer func() { stop(); workers.Wait() }()
	if recoverSummaries != nil {
		workers.Go(func() { runDelivery(ctx, "summary", 90*time.Second, recoverSummaries) })
	}
	if topic := os.Getenv("OBJECT_QUANTITY_PUBSUB_TOPIC"); topic != "" {
		if repositoryStore == nil || !regexp.MustCompile(`^projects/[a-zA-Z0-9-]+/topics/[A-Za-z][A-Za-z0-9._~%+-]{2,254}$`).MatchString(topic) {
			return errors.New("invalid Object event configuration: repository storage and full Pub/Sub topic required")
		}
		client, err := pubsub.NewService(ctx)
		if err != nil {
			return err
		}
		publisher := objectmessages.Publisher{Client: client, Topic: topic}
		store := objectevents.Store{Client: repositoryStore.Client, ChannelID: repositoryStore.ChannelID}
		workers.Go(func() {
			runDelivery(ctx, "object_quantity", 50*time.Second, func(batch context.Context) error {
				return objectcommand.DeliverQuantityEvents(batch, &store, publisher.Publish)
			})
		})
	}
	result := make(chan error, 1)
	slog.Info("server starting", "port", port)
	go func() { result <- server.ListenAndServe() }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
