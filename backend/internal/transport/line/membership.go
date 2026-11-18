package line

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

var ErrMembershipUnavailable = errors.New("group membership unavailable")

type Membership struct {
	Token  string
	Client *http.Client
}

// A historical interaction only helps discovery. Every response checks current
// membership through the Bot's group-member endpoint, without caching profiles.
func (m *Membership) IsMember(ctx context.Context, group, user string) (bool, error) {
	if m == nil || m.Client == nil || m.Token == "" || !ValidGroupID(group) || user == "" {
		return false, ErrMembershipUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.line.me/v2/bot/group/"+url.PathEscape(group)+"/member/"+url.PathEscape(user), nil)
	if err != nil {
		return false, ErrMembershipUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+m.Token)
	response, err := m.Client.Do(req)
	if err != nil {
		return false, ErrMembershipUnavailable
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if response.StatusCode != http.StatusOK {
		return false, ErrMembershipUnavailable
	}
	var profile struct {
		UserID string `json:"userId"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 32768)).Decode(&profile) != nil || profile.UserID != user {
		return false, ErrMembershipUnavailable
	}
	return true, nil
}
