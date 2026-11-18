package line

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/identity"
)

type verifierTransport func(*http.Request) (*http.Response, error)

func (f verifierTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestIdentityVerifierPreservesClaimChecks(t *testing.T) {
	for _, tc := range []struct {
		name, issuer, subject, audience string
		expire                          int64
		status                          int
		want                            error
	}{
		{"valid", "https://access.line.me", "actor", "123", time.Now().Add(time.Hour).Unix(), 200, nil},
		{"issuer", "https://invalid.example", "actor", "123", time.Now().Add(time.Hour).Unix(), 200, identity.ErrInvalidClaims},
		{"audience", "https://access.line.me", "actor", "456", time.Now().Add(time.Hour).Unix(), 200, identity.ErrInvalidClaims},
		{"missing actor", "https://access.line.me", "", "123", time.Now().Add(time.Hour).Unix(), 200, identity.ErrInvalidClaims},
		{"expired", "https://access.line.me", "actor", "123", time.Now().Add(-time.Hour).Unix(), 200, identity.ErrInvalidClaims},
		{"rejected", "", "", "", 0, 400, identity.ErrRejectedToken},
		{"unavailable", "", "", "", 0, 500, identity.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := &Verifier{ChannelID: "123", Client: &http.Client{Transport: verifierTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "api.line.me" || r.Method != "POST" {
					t.Fatal("identity endpoint changed")
				}
				if err := r.ParseForm(); err != nil || r.Form.Get("id_token") != "test-token" || r.Form.Get("client_id") != "123" {
					t.Fatal("identity request changed")
				}
				body := fmt.Sprintf(`{"iss":%q,"sub":%q,"aud":%q,"exp":%d}`, tc.issuer, tc.subject, tc.audience, tc.expire)
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			user, err := v.Verify(context.Background(), "test-token")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
			if err == nil && user.ID != "actor" {
				t.Fatal("verified actor lost")
			}
		})
	}
}
