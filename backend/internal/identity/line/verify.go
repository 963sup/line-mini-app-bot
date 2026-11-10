package line

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"
)

type Verifier struct {
	ChannelID string
	Client    *http.Client
}

func (v *Verifier) Verify(ctx context.Context, token string) (identity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	form := url.Values{"id_token": {token}, "client_id": {v.ChannelID}}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.line.me/oauth2/v2.1/verify",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return identity.User{}, identity.ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := v.Client.Do(req)
	if err != nil {
		return identity.User{}, identity.ErrUnavailable
	}
	// The response is read-only; closing cannot commit application state.
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusUnauthorized {
		return identity.User{}, identity.ErrRejectedToken
	}
	if response.StatusCode != http.StatusOK {
		return identity.User{}, identity.ErrUnavailable
	}
	var claims struct {
		Issuer   string `json:"iss"`
		Subject  string `json:"sub"`
		Audience string `json:"aud"`
		Expires  int64  `json:"exp"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 16<<10)).Decode(&claims); err != nil {
		return identity.User{}, identity.ErrInvalidResponse
	}
	validChannel := claims.Issuer == "https://access.line.me" && claims.Audience == v.ChannelID
	validUser := claims.Subject != "" && claims.Expires > time.Now().Unix()
	if !validChannel || !validUser {
		return identity.User{}, identity.ErrInvalidClaims
	}
	return identity.User{ID: claims.Subject, Name: claims.Name}, nil
}
