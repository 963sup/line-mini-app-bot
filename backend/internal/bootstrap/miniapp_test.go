package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"
	identityline "github.com/963sup/line-mini-app-bot/internal/identity/line"
	navigation "github.com/963sup/line-mini-app-bot/internal/navigation/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type miniAppRoundTrip func(*http.Request) (*http.Response, error)

func (f miniAppRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMiniAppProfiles(t *testing.T) {
	// Legacy keys cannot mix a different audience into the selected profile.
	t.Setenv("LINE_MINI_APP_CHANNEL_ID", "999")
	t.Setenv("LINE_MINI_APP_LIFF_ID", "999-legacy")
	t.Setenv("LINE_MINI_APP_WEB_URL", "https://attacker.example/miniapp")
	for _, test := range []struct{ environment, channelID, liffID string }{
		{"", "2012412007", "2012412007-0xCIzCmZ"},
		{"developing", "2012412007", "2012412007-0xCIzCmZ"},
		{"review", "2012412008", "2012412008-L7dYhvM1"},
		{"published", "2012412009", "2012412009-Xp2CQWOP"},
	} {
		t.Run(test.environment, func(t *testing.T) {
			prefix := "LINE_MINI_APP_" + strings.ToUpper(test.environment) + "_"
			if test.environment == "" {
				prefix = "LINE_MINI_APP_DEVELOPING_"
			}
			t.Setenv(prefix+"CHANNEL_ID", test.channelID)
			t.Setenv(prefix+"LIFF_ID", test.liffID)
			endpoint := "https://app.example/miniapp?view=desktop"
			t.Setenv(prefix+"ENDPOINT_URL", endpoint)
			auth, endpointURL, err := newMiniAppAuthenticator(test.environment)
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
			event := webhook.PostbackEvent{Mode: "active", ReplyToken: "reply", Source: webhook.UserSource{UserId: "user"}, Postback: &webhook.PostbackContent{Data: "action=miniapp"}}
			card := navigation.ReplyForEvent(event, link, endpointURL).Message.(*messaging_api.FlexMessage)
			action := card.Contents.(*messaging_api.FlexBubble).Footer.Contents[0].(*messaging_api.FlexButton).Action.(*messaging_api.UriAction)
			if action.Uri != link || action.AltUri == nil || action.AltUri.Desktop != endpoint {
				t.Fatal("each profile must keep its permalink and use the selected desktop endpoint")
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
		if auth, _, err := newMiniAppAuthenticator(environment); err == nil || auth != nil {
			t.Fatalf("accepted invalid environment %q", environment)
		}
	}
}

func TestMiniAppSelectedConfigurationValidation(t *testing.T) {
	const prefix = "LINE_MINI_APP_REVIEW_"
	defaults := map[string]string{"CHANNEL_ID": "123", "LIFF_ID": "123-example", "ENDPOINT_URL": "https://app.example/miniapp/nested?view=desktop&source=line"}
	for _, stage := range []string{"DEVELOPING", "PUBLISHED"} {
		for key := range defaults {
			t.Setenv("LINE_MINI_APP_"+stage+"_"+key, "")
		}
	}
	for key, value := range defaults {
		t.Setenv(prefix+key, value)
	}
	if _, endpoint, err := newMiniAppAuthenticator("review"); err != nil || endpoint != defaults["ENDPOINT_URL"] {
		t.Fatalf("unselected profiles must not be required: %v", err)
	}
	for _, test := range []struct{ key, value string }{
		{"CHANNEL_ID", ""},
		{"CHANNEL_ID", "123x"},
		{"CHANNEL_ID", " 123"},
		{"LIFF_ID", ""},
		{"LIFF_ID", "999-other"},
		{"LIFF_ID", "123-"},
		{"ENDPOINT_URL", ""},
		{"ENDPOINT_URL", "http://app.example/miniapp"},
		{"ENDPOINT_URL", "/miniapp"},
		{"ENDPOINT_URL", "https://user:password@app.example/miniapp"},
		{"ENDPOINT_URL", "https://app.example/miniapp#fragment"},
	} {
		t.Run(test.key+"/"+test.value, func(t *testing.T) {
			t.Setenv(prefix+test.key, test.value)
			auth, endpoint, err := newMiniAppAuthenticator("review")
			if err == nil || auth != nil || endpoint != "" || !strings.Contains(err.Error(), prefix+test.key) {
				t.Fatalf("invalid selected setting not rejected by key: %v", err)
			}
		})
	}
	// Existing legacy keys cannot supply a missing selected value.
	t.Setenv(prefix+"LIFF_ID", "")
	t.Setenv("LINE_MINI_APP_LIFF_ID", "123-legacy")
	if _, _, err := newMiniAppAuthenticator("review"); err == nil {
		t.Fatal("legacy fallback must not enable a profile")
	}
	// Configuration does not mutate process settings or derive a URL from a request.
	if os.Getenv(prefix+"ENDPOINT_URL") != defaults["ENDPOINT_URL"] {
		t.Fatal("configuration changed runtime settings")
	}
}
