package identityhttp

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/httpapi"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	"github.com/963sup/line-mini-app-bot/internal/identity"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type Authenticator struct {
	ChannelID, LIFFID string
	Verifier          identity.Verifier
}

func (a *Authenticator) Authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	if user, ok := identity.FromContext(r.Context()); ok {
		return user.ID, user.Name, true
	}
	if a.ChannelID == "" {
		transport.WriteAPIError(w, 503, "not_configured", "MINI App 尚未設定。")
		return "", "", false
	}
	fields := strings.Fields(r.Header.Get("Authorization"))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		transport.WriteAPIError(w, 401, "unauthenticated", "請先登入 LINE。")
		return "", "", false
	}
	if len(fields[1]) > 16384 {
		transport.WriteAPIError(w, 401, "unauthenticated", "無效的 LINE token。")
		return "", "", false
	}
	user, err := a.Verifier.Verify(r.Context(), fields[1])
	if err != nil {
		code, status, message := "unavailable", http.StatusServiceUnavailable, "暫時無法驗證 LINE 身分。"
		switch {
		case errors.Is(err, identity.ErrRejectedToken):
			code, status, message = "unauthenticated", 401, "LINE 登入已失效，請重新登入。"
		case errors.Is(err, identity.ErrInvalidClaims):
			code, status, message = "unauthenticated", 401, "LINE 身分驗證未通過。"
		case errors.Is(err, identity.ErrInvalidResponse):
			message = "LINE 驗證回應無效。"
		}
		transport.WriteAPIError(w, status, code, message)
		return "", "", false
	}
	return user.ID, user.Name, true
}

func (a *Authenticator) Register(mux *http.ServeMux) (string, error) {
	if a.ChannelID != "" || a.LIFFID != "" {
		if !regexp.MustCompile(`^[0-9]+$`).MatchString(a.ChannelID) {
			return "", fmt.Errorf("missing or invalid configuration: LINE_MINI_APP_CHANNEL_ID")
		}
		if !regexp.MustCompile(`^[0-9]+-[A-Za-z0-9]+$`).MatchString(a.LIFFID) {
			return "", fmt.Errorf("missing or invalid configuration: LINE_MINI_APP_LIFF_ID")
		}
	}
	routes := http.NewServeMux()
	routes.HandleFunc("GET /api/v1/miniapp/config", func(w http.ResponseWriter, r *http.Request) {
		if a.LIFFID == "" {
			transport.WriteAPIError(w, 503, "not_configured", "MINI App 尚未設定。")
			return
		}
		transport.WriteJSON(w, 200, generated.MiniAppConfig{LiffId: a.LIFFID})
	})
	routes.HandleFunc("GET /api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		user, name, ok := a.Authenticate(w, r)
		if ok {
			transport.WriteJSON(w, 200, generated.CurrentUser{UserId: user, DisplayName: name})
		}
	})
	httpapi.RegisterIdentity(mux, routes)
	if a.LIFFID == "" {
		return "", nil
	}
	return "https://miniapp.line.me/" + a.LIFFID, nil
}
