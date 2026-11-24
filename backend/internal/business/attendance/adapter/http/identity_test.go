package attendancehttp

import (
	"net/http"

	identityline "github.com/963sup/line-mini-app-bot/internal/platform/line"
	identityhttp "github.com/963sup/line-mini-app-bot/internal/transport/http/identity"
)

func testAuthenticator(client *http.Client) *identityhttp.Authenticator {
	return &identityhttp.Authenticator{ChannelID: "123", Verifier: &identityline.Verifier{ChannelID: "123", Client: client}}
}
