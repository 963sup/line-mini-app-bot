package line

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type membershipTransport func(*http.Request) (*http.Response, error)

func (f membershipTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMembershipRequiresMatchingCurrentProfile(t *testing.T) {
	group := "C" + strings.Repeat("a", 32)
	for _, tc := range []struct {
		name, body string
		status     int
		member     bool
		failed     bool
	}{
		{"matching profile", `{"userId":"actor"}`, 200, true, false},
		{"not a member", ``, 404, false, false},
		{"rate limited", `{}`, 429, false, true},
		{"forbidden", `{}`, 403, false, true},
		{"mismatched profile", `{"userId":"another"}`, 200, false, true},
		{"missing profile", `{}`, 200, false, true},
		{"invalid JSON", `{`, 200, false, true},
		{"oversized profile", `{"padding":"` + strings.Repeat("x", 32768) + `","userId":"actor"}`, 200, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: membershipTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodGet || r.URL.Host != "api.line.me" || r.URL.EscapedPath() != "/v2/bot/group/"+group+"/member/actor" || r.Header.Get("Authorization") != "Bearer credential" {
					t.Fatalf("unexpected membership request: %s %s", r.Method, r.URL)
				}
				deadline, ok := r.Context().Deadline()
				if !ok || time.Until(deadline) > time.Second {
					t.Fatal("membership request must remain bounded")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			membership := Membership{Token: "credential", Client: client}
			got, err := membership.IsMember(context.Background(), group, "actor")
			if got != tc.member || errors.Is(err, ErrMembershipUnavailable) != tc.failed || calls != 1 {
				t.Fatalf("member=%v err=%v calls=%d", got, err, calls)
			}
		})
	}
}

func TestMembershipFailsClosedBeforeSendingInvalidInput(t *testing.T) {
	client := &http.Client{Transport: membershipTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid input must not send a membership request")
		return nil, errors.New("unexpected request")
	})}
	group := "C" + strings.Repeat("a", 32)
	for _, tc := range []struct {
		membership  *Membership
		group, user string
	}{
		{nil, group, "actor"},
		{&Membership{Token: "credential"}, group, "actor"},
		{&Membership{Client: client}, group, "actor"},
		{&Membership{Token: "credential", Client: client}, "bad-group", "actor"},
		{&Membership{Token: "credential", Client: client}, group, ""},
	} {
		if got, err := tc.membership.IsMember(context.Background(), tc.group, tc.user); got || !errors.Is(err, ErrMembershipUnavailable) {
			t.Fatalf("invalid request allowed: member=%v err=%v", got, err)
		}
	}
}

func TestMembershipCancellationAndRedirectStayUnavailable(t *testing.T) {
	group := "C" + strings.Repeat("a", 32)
	t.Run("cancelled request", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		client := &http.Client{Transport: membershipTransport(func(r *http.Request) (*http.Response, error) {
			return nil, r.Context().Err()
		})}
		membership := Membership{Token: "credential", Client: client}
		if member, err := membership.IsMember(ctx, group, "actor"); member || !errors.Is(err, ErrMembershipUnavailable) {
			t.Fatalf("cancelled membership allowed: %v %v", member, err)
		}
	})
	t.Run("redirect rejected by configured client", func(t *testing.T) {
		calls := 0
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Transport: membershipTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": []string{"https://another.example/profile"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		})}
		membership := Membership{Token: "credential", Client: client}
		if member, err := membership.IsMember(context.Background(), group, "actor"); member || !errors.Is(err, ErrMembershipUnavailable) || calls != 1 {
			t.Fatalf("redirect membership allowed or followed: %v %v calls=%d", member, err, calls)
		}
	})
}
