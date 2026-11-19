package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"
	identityline "github.com/963sup/line-mini-app-bot/internal/identity/line"
)

type miniAppRoundTrip func(*http.Request) (*http.Response, error)

func (f miniAppRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMiniAppProfiles(t *testing.T) {
	// Legacy keys cannot mix a different audience into the selected profile.
	t.Setenv("LINE_MINI_APP_CHANNEL_ID", "999")
	t.Setenv("LINE_MINI_APP_LIFF_ID", "999-legacy")
	for _, test := range []struct{ environment, channelID, liffID string }{
		{"", "2012412007", "2012412007-0xCIzCmZ"},
		{"developing", "2012412007", "2012412007-0xCIzCmZ"},
		{"review", "2012412008", "2012412008-L7dYhvM1"},
		{"published", "2012412009", "2012412009-Xp2CQWOP"},
	} {
		t.Run(test.environment, func(t *testing.T) {
			auth, err := newMiniAppAuthenticator(test.environment)
			if err != nil {
				t.Fatal(err)
			}
			if auth.ChannelID != test.channelID || auth.LIFFID != test.liffID {
				t.Fatalf("mixed profile: channel=%s LIFF=%s", auth.ChannelID, auth.LIFFID)
			}
			mux := http.NewServeMux()
			link, err := auth.Register(mux)
			if err != nil || link != "https://miniapp.line.me/"+test.liffID {
				t.Fatalf("unexpected permanent link %q: %v", link, err)
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/miniapp/config", nil))
			var config struct {
				LIFFID string `json:"liffId"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &config); err != nil || response.Code != http.StatusOK || config.LIFFID != test.liffID {
				t.Fatalf("unexpected config: %s (%v)", response.Body, err)
			}
			verifier := auth.Verifier.(*identityline.Verifier)
			otherChannel := "2012412009"
			if test.channelID == otherChannel {
				otherChannel = "2012412007"
			}
			for _, audience := range []string{test.channelID, otherChannel} {
				verifier.Client.Transport = miniAppRoundTrip(func(r *http.Request) (*http.Response, error) {
					if err := r.ParseForm(); err != nil || r.Form.Get("client_id") != test.channelID {
						t.Fatalf("verification must use only selected channel: %v", err)
					}
					body := fmt.Sprintf(`{"iss":"https://access.line.me","sub":"user","aud":%q,"exp":%d}`, audience, time.Now().Add(time.Hour).Unix())
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				_, err := verifier.Verify(context.Background(), "test-token")
				if audience == test.channelID && err != nil {
					t.Fatal(err)
				}
				if audience != test.channelID && !errors.Is(err, identity.ErrInvalidClaims) {
					t.Fatalf("accepted another audience: %v", err)
				}
			}
		})
	}
}

func TestMiniAppProfileRejectsUnknownEnvironment(t *testing.T) {
	for _, environment := range []string{"production", "Review", " developing", "published,developing"} {
		if auth, err := newMiniAppAuthenticator(environment); err == nil || auth != nil {
			t.Fatalf("accepted invalid environment %q", environment)
		}
	}
}
