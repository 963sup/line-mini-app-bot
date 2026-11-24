package statisticshttp

import (
	repositoryApplicationStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/application/statistics"
	repositoryDomainEntityStatistics "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/statistics"
)

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/platform/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api"
	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generated"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

type (
	Authenticator interface {
		Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
	}
	statisticsServer struct {
		service *repositoryApplicationStatistics.Service
	}
)

var _ httpapi.StatisticsServer = (*statisticsServer)(nil)

func Register(mux *http.ServeMux, service *repositoryApplicationStatistics.Service, auth Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterStatistics(mux, &statisticsServer{service}, func(next http.Handler) http.Handler {
		document := *spec
		validate := middleware.OapiRequestValidatorWithOptions(&document, &middleware.Options{
			DoNotValidateServers: true,
			Options: openapi3filter.Options{AuthenticationFunc: func(ctx context.Context, _ *openapi3filter.AuthenticationInput) error {
				if _, ok := identity.FromContext(ctx); !ok {
					return errors.New("unverified identity")
				}
				return nil
			}},
			ErrorHandlerWithOpts: func(_ context.Context, _ error, w http.ResponseWriter, _ *http.Request, _ middleware.ErrorHandlerOpts) {
				transport.WriteAPIError(w, 400, "invalid_input", "請求資料不符合統計契約。")
			},
		})(next)
		return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			user, name, ok := auth.Authenticate(w, r)
			if !ok {
				return
			}
			if service == nil {
				transport.WriteAPIError(w, 503, "not_configured", "統計尚未啟用。")
				return
			}
			if r.Method == http.MethodHead {
				r = r.Clone(r.Context())
				r.Method = http.MethodGet
			}
			if !transport.ValidateUniqueQuery(w, r) {
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
			defer cancel()
			validate.ServeHTTP(w, r.WithContext(identity.WithUser(ctx, identity.User{ID: user, Name: name})))
		}))
	})
}

func (s *statisticsServer) GetRepositoryStatistics(w http.ResponseWriter, r *http.Request, repositoryId string) {
	u, _ := identity.FromContext(r.Context())
	v, e := s.service.Get(r.Context(), u.ID, repositoryId)
	if e != nil {
		if errors.Is(e, repositoryDomainEntityStatistics.ErrNotFound) {
			transport.WriteAPIError(w, 404, "not_found", "找不到此空間。")
		} else {
			transport.WriteAPIError(w, 503, "unavailable", "統計讀取失敗，請重新載入。")
		}
		return
	}
	transport.WriteJSON(w, 200, generated.RepositoryStatistics{OpenIssues: v.OpenIssues, InProgressIssues: v.InProgressIssues, CompletedIssues: v.CompletedIssues, Discussions: v.Discussions, Milestones: v.Milestones, Projects: v.Projects, ObservedAt: v.ObservedAt})
}
