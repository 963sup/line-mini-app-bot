package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	projectApplication "github.com/963sup/line-mini-app-bot/internal/business/project/application"

	"cloud.google.com/go/firestore"

	attendancehttp "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/http"
	attendanceline "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/line"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/line/presentation"
	discussionhttp "github.com/963sup/line-mini-app-bot/internal/business/discussion/adapter/http"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"
	enterprisehttp "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/http"
	issuehttp "github.com/963sup/line-mini-app-bot/internal/business/issue/adapter/http"
	journalhttp "github.com/963sup/line-mini-app-bot/internal/business/journal/adapter/http"
	membershipdb "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/firestore"
	membershiphttp "github.com/963sup/line-mini-app-bot/internal/business/membership/adapter/http"
	membershipapp "github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	objecthttp "github.com/963sup/line-mini-app-bot/internal/business/object/adapter/http"
	operationshttp "github.com/963sup/line-mini-app-bot/internal/business/operations/adapter/http"
	poppler "github.com/963sup/line-mini-app-bot/internal/business/operations/adapter/poppler"
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
	organizationdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	organizationhttp "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/http"
	organizationline "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/line"
	projecthttp "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/http"
	repositoryhttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http"
	labelhttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/labels"
	milestonehttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/milestones"
	businessgrouphttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/space"
	statisticshttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/statistics"
	wikihttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http/wiki"
	repositoryline "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/line"
	teamdb "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/firestore"
	teamhttp "github.com/963sup/line-mini-app-bot/internal/business/team/adapter/http"
	grouphttp "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/http"
	groupline "github.com/963sup/line-mini-app-bot/internal/linegroup/adapter/line"
	firestoreplatform "github.com/963sup/line-mini-app-bot/internal/platform/firestore"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/browser"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/geolocation"
)

func Run() (resultErr error) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
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
	sessions, sessionStore, err := newSessions(firestoreConfigured, getFirestore, auth.ChannelID)
	if err != nil {
		return err
	}
	auth.Sessions = sessions
	appURL, err := auth.Register(mux)
	if err != nil {
		return err
	}
	attendance, err := newAttendance(firestoreConfigured, getFirestore, os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"))
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
		membership := &lineplatform.Membership{Timeout: 5 * time.Second, Token: os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN"), Client: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
		browserAttendance = attendancehttp.NewAttendance(attendance.store, attendance.zone, membership, func(ctx context.Context, group string) string {
			return presentation.GroupName(group, lookup(ctx, group))
		})
	}

	attendancehttp.RegisterAttendanceAPI(mux, browserAttendance, auth)
	operationshttp.Register(mux, &operations.DocumentTextService{Extractor: &poppler.Extractor{}}, auth)
	summaries, err := registerSummary(context.Background(), mux, firestoreConfigured, getFirestore)
	if err != nil {
		return err
	}
	repositories, repositoryStore, err := newRepositories(firestoreConfigured, getFirestore)
	if err != nil {
		return err
	}
	groups, groupStore, err := newGroups(firestoreConfigured, getFirestore)
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
	if sessionStore != nil && os.Getenv("LINE_MESSAGING_API_CHANNEL_ID") != "" {
		channel := os.Getenv("LINE_MESSAGING_API_CHANNEL_ID")
		organizationStore := organizationdb.Store{
			DB: sessionStore.Client, Channel: channel,
			EnterpriseNavigation: enterprisedb.NavigationFacts,
			CanCreateEnterprise:  enterprisedb.CanStartCreation,
			CreateEnterprise:     enterprisedb.CreateFromConversation,
		}
		organizationHandler.Store = organizationStore
		enterpriseStore := enterprisedb.Store{
			DB: sessionStore.Client, Channel: channel,
			OrganizationMember: membershipdb.OrganizationMember,
			TeamMember:         membershipdb.TeamMember,
		}
		teamStore := teamdb.Store{
			DB: sessionStore.Client, Channel: channel,
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
		invitations = membershipInvitations(sessionStore.Client, channel, os.Getenv("LINE_MESSAGING_API_CHANNEL_SECRET"))
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
	if sessionStore != nil {
		projects = newProjects(sessionStore.Client, os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"), repositories)
	}
	journalhttp.Register(mux, newJournals(groupCatalog, groupStore, repositories, projects), auth)
	grouphttp.Register(mux, groupCatalog, auth)
	repositoryhttp.Register(mux, repositories, auth)
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
		repositoryObserver = &repositoryline.Observer{Store: repositories.Store}
	}
	var groupObserver *groupline.Observer
	if lineTeams != nil {
		groupObserver = &groupline.Observer{Store: lineTeams}
	}
	groupLookup, err = registerLINE(mux, appURL, endpointURL, punches, summaries, repositoryObserver, groupObserver, repositories, creation, invitations)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr: net.JoinHostPort("", port), Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 100 * time.Second, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
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
