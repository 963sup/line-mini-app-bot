package server

import (
	"fmt"
	"net/http"
	"time"

	identityhttp "github.com/963sup/line-mini-app-bot/internal/identity/http"
	identityline "github.com/963sup/line-mini-app-bot/internal/identity/line"
)

// Public channel/LIFF pairs are selected together; tokens are never accepted
// against another environment's channel.
func newMiniAppAuthenticator(environment string) (*identityhttp.Authenticator, error) {
	var channelID, liffID string
	switch environment {
	case "", "developing":
		channelID, liffID = "2012412007", "2012412007-0xCIzCmZ"
	case "review":
		channelID, liffID = "2012412008", "2012412008-L7dYhvM1"
	case "published":
		channelID, liffID = "2012412009", "2012412009-Xp2CQWOP"
	default:
		return nil, fmt.Errorf("invalid LINE_MINI_APP_ENV %q: use developing, review or published", environment)
	}
	return &identityhttp.Authenticator{
		ChannelID: channelID,
		LIFFID:    liffID,
		Verifier: &identityline.Verifier{
			ChannelID: channelID,
			Client:    &http.Client{Timeout: 5 * time.Second},
		},
	}, nil
}
