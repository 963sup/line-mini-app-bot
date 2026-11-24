package bootstrap

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	identityline "github.com/963sup/line-mini-app-bot/internal/platform/line"
	identityhttp "github.com/963sup/line-mini-app-bot/internal/transport/http/identity"
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
	if !validChannelID(channelID) {
		return nil, "", fmt.Errorf("%sCHANNEL_ID must contain only digits", prefix)
	}
	if !strings.HasPrefix(liffID, channelID+"-") || !regexp.MustCompile(`^[0-9]+-[A-Za-z0-9]+$`).MatchString(liffID) {
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
