package httpadapter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (app *miniApp) handleMe(w http.ResponseWriter, r *http.Request) {
	user, name, ok := app.authenticate(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, struct {
		UserID      string `json:"userId"`
		DisplayName string `json:"displayName"`
	}{UserID: user, DisplayName: name})
}

func (app *miniApp) authenticate(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	if app.channelID == "" {
		writeAPIError(w, http.StatusServiceUnavailable, "not_configured", "MINI App 尚未設定。")
		return "", "", false
	}
	authorization := strings.Fields(r.Header.Get("Authorization"))
	if len(authorization) != 2 || !strings.EqualFold(authorization[0], "Bearer") {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "請先登入 LINE。")
		return "", "", false
	}
	if len(authorization[1]) > 16384 {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "無效的 LINE token。")
		return "", "", false
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	form := url.Values{"id_token": {authorization[1]}, "client_id": {app.channelID}}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.line.me/oauth2/v2.1/verify",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "暫時無法驗證 LINE 身分。")
		return "", "", false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.client.Do(req)
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "暫時無法驗證 LINE 身分。")
		return "", "", false
	}
	// The response is read-only; closing cannot commit application state.
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusUnauthorized {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "LINE 登入已失效，請重新登入。")
		return "", "", false
	}
	if response.StatusCode != http.StatusOK {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "暫時無法驗證 LINE 身分。")
		return "", "", false
	}
	var claims struct {
		Issuer   string `json:"iss"`
		Subject  string `json:"sub"`
		Audience string `json:"aud"`
		Expires  int64  `json:"exp"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 16<<10)).Decode(&claims); err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "unavailable", "LINE 驗證回應無效。")
		return "", "", false
	}
	validChannel := claims.Issuer == "https://access.line.me" && claims.Audience == app.channelID
	validUser := claims.Subject != "" && claims.Expires > time.Now().Unix()
	if !validChannel || !validUser {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "LINE 身分驗證未通過。")
		return "", "", false
	}
	return claims.Subject, claims.Name, true
}
