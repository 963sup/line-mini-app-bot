package enterprisehttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	enterprise "github.com/963sup/line-mini-app-bot/internal/business/enterprise/domain"
	orghttp "github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/http"
	legacy "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
)

type responseAuth struct{}

func (responseAuth) Authenticate(http.ResponseWriter, *http.Request) (string, string, bool) {
	return "actor", "", true
}

func TestEnterpriseErrorsPreservePublicHTTPContract(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"enterprise missing", enterprise.ErrNotFound, 404, "not_found"},
		{"enterprise conflict", enterprise.ErrConflict, 409, "conflict"},
		{"legacy storage missing", legacy.ErrNotFound, 404, "not_found"},
		{"legacy provisioning denial", legacy.ErrForbidden, 403, "forbidden"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/enterprises/enterprise/organizations", nil)
			orghttp.Serve(response, request, responseAuth{}, true, func(context.Context, string) (any, error) {
				return enterpriseLinkResponse(nil, fmt.Errorf("wrapped: %w", tc.err))
			})
			var body struct{ Error struct{ Code string } }
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != tc.status || body.Error.Code != tc.code {
				t.Fatalf("contract changed: %d %s", response.Code, response.Body)
			}
		})
	}
}
