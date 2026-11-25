package httpapi

import (
	"context"
	"net/http"
	"strings"
	"sync"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
	"github.com/getkin/kin-openapi/openapi3"
)

var Spec = sync.OnceValues(func() (*openapi3.T, error) {
	spec, err := generated.GetSpec()
	if err != nil {
		return nil, err
	}
	if err := spec.Validate(context.Background()); err != nil {
		return nil, err
	}
	return spec, nil
})

func routeHandler(server contractServer) http.Handler {
	return generated.HandlerWithOptions(server, generated.StdHTTPServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
		transport.WriteAPIError(w, 400, "invalid_input", "請求參數無效。")
	}})
}

func withMiddleware(handler http.Handler, wrap func(http.Handler) http.Handler) http.Handler {
	if wrap != nil {
		return wrap(handler)
	}
	return handler
}

func RegisterIdentity(public *http.ServeMux, handler IdentityServer) {
	routes := routeHandler(contractServer{IdentityServer: handler})
	public.Handle("GET /api/v1/miniapp/config", routes)
	public.Handle("GET /api/v1/me", routes)
	public.Handle("POST /api/v1/session", routes)
	public.Handle("POST /api/v1/session/refresh", routes)
	public.Handle("DELETE /api/v1/session", routes)
}

func RegisterRepositories(public *http.ServeMux, handler RepositoryServer) {
	routes := routeHandler(contractServer{RepositoryServer: handler})
	public.Handle("GET /api/v1/repositories", routes)
	public.Handle("GET /api/v1/repositories/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{id}", routes)
}

func RegisterExplore(public *http.ServeMux, handler ExploreServer) {
	routes := routeHandler(contractServer{ExploreServer: handler})
	public.Handle("GET /api/v1/explore", routes)
}

func RegisterLabels(public *http.ServeMux, handler LabelServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{LabelServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/labels", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/labels", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/labels/{id}", routes)
}

func RegisterIssues(public *http.ServeMux, handler IssueServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{IssueServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/issues", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/issues", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/issues/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/issues/{id}", routes)
}

func RegisterDiscussions(public *http.ServeMux, handler DiscussionServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{DiscussionServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/discussions", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/discussions", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/discussions/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/discussions/{id}", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/discussions/{id}/comments", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/discussions/{id}/comments", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/discussions/{id}/comments/{commentId}", routes)
}

func RegisterAttendance(public *http.ServeMux, handler AttendanceServer, wrap func(http.Handler) http.Handler) {
	spec, err := Spec()
	if err != nil {
		panic(err)
	}
	routes := withMiddleware(routeHandler(contractServer{AttendanceServer: handler}), wrap)
	for path, item := range spec.Paths.Map() {
		if !strings.HasPrefix(path, "/api/v1/attendance/") {
			continue
		}
		for method := range item.Operations() {
			public.Handle(method+" "+path, routes)
		}
	}
}

func RegisterMilestones(public *http.ServeMux, handler MilestoneServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{MilestoneServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/milestones", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/milestones", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/milestones/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/milestones/{id}", routes)
}

func RegisterProjects(public *http.ServeMux, handler ProjectServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{ProjectServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/projects", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/projects", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/projects/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/projects/{id}", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/projects/{id}/items", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/projects/{id}/items", routes)
	public.Handle("DELETE /api/v1/repositories/{repositoryId}/projects/{id}/items/{itemId}", routes)
}

func RegisterStatistics(public *http.ServeMux, handler StatisticsServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{StatisticsServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/statistics", routes)
}

func RegisterGroups(public *http.ServeMux, handler GroupServer) {
	routes := routeHandler(contractServer{GroupServer: handler})
	public.Handle("GET /api/v1/groups", routes)
	public.Handle("GET /api/v1/groups/{id}", routes)
}

func RegisterOwnerProjects(public *http.ServeMux, handler OwnerProjectServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{OwnerProjectServer: handler}), wrap)
	public.Handle("GET /api/v1/owners/{ownerId}/projects", routes)
	public.Handle("POST /api/v1/owners/{ownerId}/projects", routes)
	public.Handle("GET /api/v1/owners/{ownerId}/projects/{id}", routes)
	public.Handle("PATCH /api/v1/owners/{ownerId}/projects/{id}", routes)
	public.Handle("GET /api/v1/owners/{ownerId}/projects/{id}/items", routes)
	public.Handle("POST /api/v1/owners/{ownerId}/projects/{id}/items", routes)
	public.Handle("DELETE /api/v1/owners/{ownerId}/projects/{id}/items/{itemId}", routes)
	public.Handle("GET /api/v1/owners/{ownerId}/projects/{id}/references/{itemId}/inspection-submissions", routes)
	public.Handle("POST /api/v1/owners/{ownerId}/projects/{id}/references/{itemId}/inspection-submissions", routes)
	public.Handle("GET /api/v1/owners/{ownerId}/projects/{id}/references", routes)
	public.Handle("GET /api/v1/owners/{ownerId}/projects/{id}/reference-state", routes)
	public.Handle("POST /api/v1/owners/{ownerId}/projects/{id}/references", routes)
	public.Handle("DELETE /api/v1/owners/{ownerId}/projects/{id}/references/{itemId}", routes)
}

func RegisterObjects(public *http.ServeMux, handler ObjectServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{ObjectServer: handler}), wrap)
	public.Handle("POST /api/v1/object-template-drafts", routes)
	public.Handle("GET /api/v1/object-templates", routes)
	public.Handle("POST /api/v1/object-templates", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/object-templates", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/object-templates", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/object-templates/{id}", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/object-templates/{id}/instances", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/objects/candidates", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/object-quantity-changes", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/objects", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/objects", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/objects/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/objects/{id}", routes)
}

func RegisterRepositorySpaces(public *http.ServeMux, handler RepositorySpaceServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{RepositorySpaceServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/spaces", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/spaces", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/spaces/{id}", routes)
}

func RegisterJournals(public *http.ServeMux, handler JournalServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{JournalServer: handler})
	routes = wrap(routes)
	public.Handle("GET /api/v1/journals", routes)
	public.Handle("POST /api/v1/journals", routes)
	public.Handle("GET /api/v1/journals/{journalId}", routes)
}

func RegisterWiki(public *http.ServeMux, handler WikiServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{WikiServer: handler}), wrap)
	public.Handle("GET /api/v1/repositories/{repositoryId}/wiki-pages", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/wiki-pages", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/wiki-pages/{id}", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/wiki-pages/{id}/revisions", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/wiki-pages/{id}", routes)
}

func RegisterOrganizations(public *http.ServeMux, handler OrganizationServer) {
	routes := routeHandler(contractServer{OrganizationServer: handler})
	public.Handle("GET /api/v1/me/organization-context", routes)
	public.Handle("PUT /api/v1/me/organization-context", routes)
	public.Handle("POST /api/v1/organizations", routes)
	public.Handle("PUT /api/v1/organizations/{organizationId}/owner", routes)
}

func RegisterMembership(public *http.ServeMux, handler MembershipServer) {
	routes := routeHandler(contractServer{MembershipServer: handler})
	public.Handle("GET /api/v1/organizations/{organizationId}/members", routes)
	public.Handle("DELETE /api/v1/organizations/{organizationId}/membership", routes)
	public.Handle("DELETE /api/v1/organizations/{organizationId}/members/{memberId}", routes)
}

func RegisterTeams(public *http.ServeMux, handler TeamServer) {
	routes := routeHandler(contractServer{TeamServer: handler})
	public.Handle("PUT /api/v1/teams/{teamId}/management", routes)
	public.Handle("GET /api/v1/organizations/{organizationId}/teams", routes)
	public.Handle("PUT /api/v1/organizations/{organizationId}/enterprise-teams/{teamId}", routes)
	public.Handle("DELETE /api/v1/organizations/{organizationId}/enterprise-teams/{teamId}", routes)
}

func RegisterEnterprises(public *http.ServeMux, handler EnterpriseServer) {
	routes := routeHandler(contractServer{EnterpriseServer: handler})
	public.Handle("POST /api/v1/enterprises", routes)
	public.Handle("GET /api/v1/enterprises/{enterpriseId}", routes)
	public.Handle("POST /api/v1/enterprises/{enterpriseId}/owners", routes)
	public.Handle("POST /api/v1/enterprises/{enterpriseId}/organizations", routes)
}

func RegisterPoppler(public *http.ServeMux, handler PopplerServer, wrap func(http.Handler) http.Handler) {
	routes := withMiddleware(routeHandler(contractServer{PopplerServer: handler}), wrap)
	public.Handle("POST /api/v1/tools/poppler/extractions", routes)
}
