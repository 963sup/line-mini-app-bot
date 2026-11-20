package identityhttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/identity"
)

type rejectedVerifier struct{ err error }

func (v rejectedVerifier) Verify(context.Context, string) (identity.User, error) {
	return identity.User{}, v.err
}

func TestRejectedTokenMessageDoesNotExposeProviderDetail(t *testing.T) {
	for _, tc := range []struct {
		err        error
		diagnostic string
	}{
		{&identity.TokenRejection{HTTPStatus: 400}, "LINE HTTP 400"},
		{&identity.TokenRejection{HTTPStatus: 401}, "LINE HTTP 401"},
		{fmt.Errorf("%w: private@example.com private-token", identity.ErrRejectedToken), ""},
	} {
		a := Authenticator{ChannelID: "123", Verifier: rejectedVerifier{tc.err}}
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("Authorization", "Bearer private-token")
		res := httptest.NewRecorder()
		a.GetCurrentUser(res, req)
		if res.Code != 401 || strings.Contains(res.Body.String(), "private") || strings.Contains(res.Body.String(), "identity token rejected") {
			t.Fatalf("unsafe rejection response: %s", res.Body.String())
		}
		if tc.diagnostic != "" && !strings.Contains(res.Body.String(), tc.diagnostic) {
			t.Fatal("safe provider status lost")
		}
	}
}
