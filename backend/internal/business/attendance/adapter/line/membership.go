package line

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type Membership struct {
	Token  string
	Client *http.Client
}

func (m Membership) IsMember(ctx context.Context, group, user string) (bool, error) {
	switch checkGroupMember(ctx, m.Client, "https://api.line.me", m.Token, group, user) {
	case http.StatusOK:
		return true, nil
	case http.StatusBadRequest, http.StatusNotFound:
		return false, nil
	default:
		return false, errors.New("attendance membership unavailable")
	}
}

// A 200 from the Messaging API is the only positive membership evidence.
func checkGroupMember(ctx context.Context, client *http.Client, baseURL, token, group, user string) int {
	lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	endpoint := baseURL + "/v2/bot/group/" + url.PathEscape(group) + "/member/" + url.PathEscape(user)
	req, err := http.NewRequestWithContext(lookupCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	// Closing this read-only response cannot commit application state.
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}
