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
	routes := routeHandler(contractServer{identity: handler})
	public.Handle("GET /api/v1/miniapp/config", routes)
	public.Handle("GET /api/v1/me", routes)
	public.Handle("POST /api/v1/session", routes)
	public.Handle("POST /api/v1/session/refresh", routes)
	public.Handle("DELETE /api/v1/session", routes)
}

func RegisterRepositories(public *http.ServeMux, handler RepositoryServer) {
	routes := routeHandler(contractServer{repositories: handler})
	public.Handle("GET /api/v1/repositories", routes)
	public.Handle("GET /api/v1/repositories/{id}", routes)
}

func RegisterAttendance(public *http.ServeMux, handler AttendanceServer, wrap func(http.Handler) http.Handler) {
	spec, err := Spec()
	if err != nil {
		panic(fmt.Errorf("invalid generated API contract: %w", err))
	}
	routes := routeHandler(contractServer{attendance: handler})
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
