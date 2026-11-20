package bootstrap

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	identityhttp "github.com/963sup/line-mini-app-bot/internal/identity/http"
	identityline "github.com/963sup/line-mini-app-bot/internal/identity/line"
)

// Only the selected environment supplies the public configuration and audience.
func newMiniAppAuthenticator(environment string) (*identityhttp.Authenticator, string, error) {
	if environment == "" {
		environment = "developing"
	}
	switch environment {
	case "developing", "review", "published":
	default:
		return nil, "", fmt.Errorf("invalid LINE_MINI_APP_ENV: use developing, review or published")
	}
	prefix := "LINE_MINI_APP_" + strings.ToUpper(environment) + "_"
	channelID, liffID, endpoint := os.Getenv(prefix+"CHANNEL_ID"), os.Getenv(prefix+"LIFF_ID"), os.Getenv(prefix+"ENDPOINT_URL")
	if channelID == "" || strings.IndexFunc(channelID, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return nil, "", fmt.Errorf("%sCHANNEL_ID must contain only digits", prefix)
	}
	if !strings.HasPrefix(liffID, channelID+"-") || len(liffID) <= len(channelID)+1 || strings.TrimSpace(liffID) != liffID {
		return nil, "", fmt.Errorf("%sLIFF_ID must be nonempty and match the selected Channel ID", prefix)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, "", fmt.Errorf("%sENDPOINT_URL must be an absolute HTTPS URL without credentials or fragment", prefix)
	}
	return &identityhttp.Authenticator{
		ChannelID: channelID,
		LIFFID:    liffID,
		Verifier: &identityline.Verifier{
			ChannelID: channelID,
			Client:    &http.Client{Timeout: 5 * time.Second},
		},
	}, endpoint, nil
}
