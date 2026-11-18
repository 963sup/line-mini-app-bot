package httpapi

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
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
}

func RegisterLabels(public *http.ServeMux, handler LabelServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{LabelServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/repositories/{repositoryId}/labels", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/labels", routes)
}

func RegisterIssues(public *http.ServeMux, handler IssueServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{IssueServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/repositories/{repositoryId}/issues", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/issues", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/issues/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/issues/{id}", routes)
}

func RegisterDiscussions(public *http.ServeMux, handler DiscussionServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{DiscussionServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
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
	routes := routeHandler(contractServer{AttendanceServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
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
	routes := routeHandler(contractServer{MilestoneServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/repositories/{repositoryId}/milestones", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/milestones", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/milestones/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/milestones/{id}", routes)
}

func RegisterProjects(public *http.ServeMux, handler ProjectServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{ProjectServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/repositories/{repositoryId}/projects", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/projects", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/projects/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/projects/{id}", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/projects/{id}/items", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/projects/{id}/items", routes)
	public.Handle("DELETE /api/v1/repositories/{repositoryId}/projects/{id}/items/{itemId}", routes)
}

func RegisterStatistics(public *http.ServeMux, handler StatisticsServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{StatisticsServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/repositories/{repositoryId}/statistics", routes)
}

func RegisterGroups(public *http.ServeMux, handler GroupServer) {
	routes := routeHandler(contractServer{GroupServer: handler})
	public.Handle("GET /api/v1/groups", routes)
	public.Handle("GET /api/v1/groups/{id}", routes)
}

func RegisterOwnerProjects(public *http.ServeMux, handler OwnerProjectServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{OwnerProjectServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
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
	public.Handle("POST /api/v1/owners/{ownerId}/projects/{id}/references", routes)
	public.Handle("DELETE /api/v1/owners/{ownerId}/projects/{id}/references/{itemId}", routes)
}

func RegisterObjects(public *http.ServeMux, handler ObjectServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{ObjectServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/repositories/{repositoryId}/objects", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/objects", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/objects/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/objects/{id}", routes)
}

func RegisterSubobjects(public *http.ServeMux, handler SubobjectServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{SubobjectServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/repositories/{repositoryId}/objects/{objectId}/subobjects", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/objects/{objectId}/subobjects", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/objects/{objectId}/subobjects/{id}", routes)
	public.Handle("PATCH /api/v1/repositories/{repositoryId}/objects/{objectId}/subobjects/{id}", routes)
}

func RegisterObjectImports(public *http.ServeMux, handler ObjectImportServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{ObjectImportServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("POST /api/v1/repositories/{repositoryId}/object-import-runs", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/object-import-runs/{id}", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/object-imports", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/object-imports/parse", routes)
	public.Handle("GET /api/v1/repositories/{repositoryId}/object-imports/{id}", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/object-imports/{id}/review", routes)
	public.Handle("POST /api/v1/repositories/{repositoryId}/object-imports/{id}/confirm", routes)
}

func RegisterRepositorySpaces(public *http.ServeMux, handler RepositorySpaceServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{RepositorySpaceServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
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
