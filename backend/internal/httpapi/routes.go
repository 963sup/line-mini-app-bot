package httpapi

import (
	"context"
	"fmt"
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
		panic(fmt.Errorf("invalid generated API contract: %w", err))
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

func RegisterSpaces(public *http.ServeMux, handler SpaceServer) {
	routes := routeHandler(contractServer{SpaceServer: handler})
	public.Handle("GET /api/v1/spaces", routes)
	public.Handle("GET /api/v1/spaces/{id}", routes)
}

func RegisterSpaceProjects(public *http.ServeMux, handler SpaceProjectServer, wrap func(http.Handler) http.Handler) {
	routes := routeHandler(contractServer{SpaceProjectServer: handler})
	if wrap != nil {
		routes = wrap(routes)
	}
	public.Handle("GET /api/v1/spaces/{spaceId}/projects", routes)
	public.Handle("POST /api/v1/spaces/{spaceId}/projects", routes)
	public.Handle("GET /api/v1/spaces/{spaceId}/projects/{id}", routes)
	public.Handle("PATCH /api/v1/spaces/{spaceId}/projects/{id}", routes)
	public.Handle("GET /api/v1/spaces/{spaceId}/projects/{id}/items", routes)
	public.Handle("POST /api/v1/spaces/{spaceId}/projects/{id}/items", routes)
	public.Handle("DELETE /api/v1/spaces/{spaceId}/projects/{id}/items/{itemId}", routes)
}
