package identityhttp

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	"github.com/963sup/line-mini-app-bot/internal/identity/session"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

const sessionCookie = "__Host-session"

func (a *Authenticator) sessionReady(w http.ResponseWriter) bool {
	w.Header().Set("Cache-Control", "no-store")
	if a.Sessions == nil {
		transport.WriteAPIError(w, 503, "not_configured", "登入服務尚未設定。")
		return false
	}
	return true
}

func sessionOrigin(w http.ResponseWriter, r *http.Request) bool {
	// Cloud Run terminates TLS before Go. Never trust X-Forwarded-Host or arbitrary origins.
	if r.Header.Get("Origin") != "https://"+r.Host || http.NewCrossOriginProtection().Check(r) != nil {
		transport.WriteAPIError(w, 403, "forbidden", "請從原網站操作。")
		return false
	}
	return true
}

func (a *Authenticator) sessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, session.ErrCSRF) {
		transport.WriteAPIError(w, 403, "forbidden", "請重新驗證後再操作。")
		return
	}
	if errors.Is(err, session.ErrInvalid) {
		transport.WriteAPIError(w, 401, "unauthenticated", "登入已失效，請重新驗證。")
		return
	}
	transport.WriteAPIError(w, 503, "unavailable", "暫時無法讀取登入狀態。")
}

func (a *Authenticator) authenticateSession(w http.ResponseWriter, r *http.Request) (session.Session, bool) {
	if !a.sessionReady(w) {
		return session.Session{}, false
	}
	cookies := r.CookiesNamed(sessionCookie)
	if len(cookies) != 1 {
		a.sessionError(w, session.ErrInvalid)
		return session.Session{}, false
	}
	value, err := a.Sessions.Load(r.Context(), cookies[0].Value)
	if err != nil {
		a.sessionError(w, err)
		return session.Session{}, false
	}
	unsafe := r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions
	if unsafe && !sessionOrigin(w, r) {
		return session.Session{}, false
	}
	tokens := r.Header.Values("X-CSRF-Token")
	// The frontend sends its session-bound token on reads too, preventing stale tabs
	// from displaying a different account's data after another tab replaces the cookie.
	if unsafe || len(tokens) > 0 {
		if len(tokens) != 1 || !value.CheckCSRF(tokens[0]) {
			transport.WriteAPIError(w, 403, "forbidden", "請重新驗證後再操作。")
			return session.Session{}, false
		}
	}
	return value, true
}

func (a *Authenticator) createSession(w http.ResponseWriter, r *http.Request) {
	if !a.sessionReady(w) || !sessionOrigin(w, r) {
		return
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
		return
	}
	body, err := transportSessionBody(w, r)
	if err != nil || !body {
		transport.WriteAPIError(w, 400, "invalid_input", "請求資料無效。")
		return
	}
	user, ok := a.verifyBearer(w, r)
	if !ok {
		return
	}
	previous := ""
	cookies := r.CookiesNamed(sessionCookie)
	if len(cookies) > 1 {
		a.sessionError(w, session.ErrInvalid)
		return
	}
	if len(cookies) == 1 {
		previous = cookies[0].Value
	}
	value, token, csrf, err := a.Sessions.Create(r.Context(), user, previous)
	if err != nil {
		a.sessionError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", Secure: true, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Expires: value.ExpiresAt, MaxAge: max(1, int(time.Until(value.ExpiresAt).Seconds())),
	})
	transport.WriteJSON(w, 200, generated.SessionResponse{
		User:      generated.CurrentUser{UserId: user.ID, DisplayName: user.Name},
		CsrfToken: csrf, ExpiresAt: value.ExpiresAt,
	})
}

func transportSessionBody(w http.ResponseWriter, r *http.Request) (bool, error) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	var value map[string]json.RawMessage
	if err := decoder.Decode(&value); err != nil {
		return false, err
	}
	if value == nil || len(value) != 0 {
		return false, nil
	}
	var trailing any
	// Require exactly one JSON object.
	err := decoder.Decode(&trailing)
	return errors.Is(err, io.EOF), nil
}

func (a *Authenticator) deleteSession(w http.ResponseWriter, r *http.Request) {
	if !a.sessionReady(w) || !sessionOrigin(w, r) {
		return
	}
	cookies := r.CookiesNamed(sessionCookie)
	if len(cookies) == 0 {
		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Expired but not yet garbage-collected sessions can still be explicitly revoked.
	if len(cookies) != 1 || !session.ValidToken(cookies[0].Value) {
		a.sessionError(w, session.ErrInvalid)
		return
	}
	value, err := a.Sessions.Store.Get(r.Context(), session.Key(cookies[0].Value))
	if errors.Is(err, session.ErrInvalid) {
		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		a.sessionError(w, err)
		return
	}
	if value.ChannelID != a.ChannelID || !value.CheckCSRF(r.Header.Get("X-CSRF-Token")) {
		transport.WriteAPIError(w, 403, "forbidden", "請重新驗證後再操作。")
		return
	}
	if err := a.Sessions.Revoke(r.Context(), cookies[0].Value); err != nil {
		a.sessionError(w, err)
		return
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", Secure: true, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Expires: time.Unix(1, 0), MaxAge: -1,
	})
}

func (a *Authenticator) refreshSession(w http.ResponseWriter, r *http.Request) {
	if !a.sessionReady(w) || !sessionOrigin(w, r) {
		return
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		transport.WriteAPIError(w, 415, "invalid_content_type", "請使用 JSON。")
		return
	}
	valid, err := transportSessionBody(w, r)
	if err != nil || !valid {
		transport.WriteAPIError(w, 400, "invalid_input", "請求資料無效。")
		return
	}
	cookies, tokens := r.CookiesNamed(sessionCookie), r.Header.Values("X-CSRF-Token")
	if len(cookies) != 1 {
		a.sessionError(w, session.ErrInvalid)
		return
	}
	if len(tokens) != 1 {
		a.sessionError(w, session.ErrCSRF)
		return
	}
	value, err := a.Sessions.Refresh(r.Context(), cookies[0].Value, tokens[0])
	if err != nil {
		a.sessionError(w, err)
		return
	}
	// Commit first; never emit cookies from the retryable transaction callback.
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: cookies[0].Value, Path: "/", Secure: true, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Expires: value.ExpiresAt, MaxAge: max(1, int(time.Until(value.ExpiresAt).Seconds())),
	})
	transport.WriteJSON(w, 200, generated.SessionResponse{
		User:      generated.CurrentUser{UserId: value.User.ID, DisplayName: value.User.Name},
		CsrfToken: tokens[0], ExpiresAt: value.ExpiresAt,
	})
}
