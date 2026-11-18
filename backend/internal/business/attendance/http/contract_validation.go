package attendancehttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/httpapi"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

func registerContractAttendance(public *http.ServeMux, server httpapi.AttendanceServer, bot *Attendance, app Authenticator) {
	spec, err := httpapi.Spec()
	if err != nil {
		panic(err)
	}
	httpapi.RegisterAttendance(public, server, func(next http.Handler) http.Handler {
		if bot == nil {
			// Disabled capabilities return 503 before generated required-parameter binding.
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				message := "出勤尚未啟用。"
				if strings.HasSuffix(r.URL.Path, "/leaves") {
					message = "請假尚未啟用。"
				}
				transport.WriteAPIError(w, 503, "not_configured", message)
			})
		}
		return validateAttendanceContract(spec, next, bot, app)
	})
}

func validateAttendanceContract(spec *openapi3.T, next http.Handler, bot *Attendance, app Authenticator) http.Handler {
	// The document is shared and immutable; middleware may clear Servers on its own copy.
	document := *spec
	validate := middleware.OapiRequestValidatorWithOptions(&document, &middleware.Options{
		DoNotValidateServers: true,
		Options: openapi3filter.Options{AuthenticationFunc: func(ctx context.Context, _ *openapi3filter.AuthenticationInput) error {
			if _, ok := identity.FromContext(ctx); !ok {
				return errors.New("identity was not verified")
			}
			return nil
		}},
		ErrorHandlerWithOpts: contractValidationError,
	})(next)
	return http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bot == nil {
			next.ServeHTTP(w, r)
			return
		}
		user, name, ok := app.Authenticate(w, r)
		if !ok {
			return
		}
		// HEAD follows the existing GET route semantics without adding an OpenAPI operation.
		if r.Method == http.MethodHead {
			r = r.Clone(r.Context())
			r.Method = http.MethodGet
		}
		// Legacy optional empty pagination values mean "use defaults".
		query := r.URL.Query()
		changed := false
		for _, name := range []string{"pageSize", "cursor"} {
			if values, exists := query[name]; exists && len(values) > 0 && values[0] == "" {
				query.Del(name)
				changed = true
			}
		}
		if changed {
			r.URL.RawQuery = query.Encode()
		}
		// Preserve existing media and payload limits before the validator reads the body.
		if r.Method == http.MethodPost && !strings.HasSuffix(r.URL.Path, "/approve") && !strings.HasSuffix(r.URL.Path, "/cancel") {
			media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || media != "application/json" {
				transport.WriteAPIError(w, http.StatusUnsupportedMediaType, "invalid_content_type", "請使用 JSON。")
				return
			}
			limit := int64(16384)
			if strings.HasSuffix(r.URL.Path, "/launch") || strings.HasSuffix(r.URL.Path, "/preflight") {
				limit = 512
			}
			if strings.HasSuffix(r.URL.Path, "/punches") {
				limit = 2048
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
		}
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/attendance/corrections" && len(r.Header.Values("Idempotency-Key")) > 1 {
			transport.WriteAPIError(w, http.StatusBadRequest, "invalid_idempotency_key", "重試識別碼無效。")
			return
		}
		validate.ServeHTTP(w, r.WithContext(identity.WithUser(r.Context(), identity.User{ID: user, Name: name})))
	}))
}

func contractValidationError(_ context.Context, err error, w http.ResponseWriter, r *http.Request, _ middleware.ErrorHandlerOpts) {
	code := "invalid_input"
	var requestError *openapi3filter.RequestError
	if errors.As(err, &requestError) && requestError.Parameter != nil {
		switch requestError.Parameter.Name {
		case "cursor":
			code = "invalid_cursor"
		case "Idempotency-Key":
			code = "invalid_idempotency_key"
		}
	}
	if strings.HasSuffix(r.URL.Path, "/launch") {
		code = "invalid_launch"
	}
	var schemaError *openapi3.SchemaError
	if strings.HasSuffix(r.URL.Path, "/punches") && errors.As(err, &schemaError) && schemaError.SchemaField != "additionalProperties" {
		code = "invalid_position"
	}
	// Never return validator details: they can contain token, location or proposal data.
	transport.WriteAPIError(w, http.StatusBadRequest, code, "請求資料不符合 API 契約，請確認內容。")
}
