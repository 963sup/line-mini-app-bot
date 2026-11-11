package attendancehttp

import (
	"net/http"

	identityhttp "github.com/963sup/line-mini-app-bot/internal/identity/http"
	identityline "github.com/963sup/line-mini-app-bot/internal/identity/line"
)

func testAuthenticator(client *http.Client) *identityhttp.Authenticator {
	return &identityhttp.Authenticator{ChannelID: "123", Verifier: &identityline.Verifier{ChannelID: "123", Client: client}}
}
