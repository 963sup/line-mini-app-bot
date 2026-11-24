package bootstrap

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	httptransport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

type messagingConfig struct{ channel, secret, token string }

func messagingEnvironment() messagingConfig {
	return messagingConfig{
		channel: os.Getenv("LINE_MESSAGING_API_CHANNEL_ID"),
		secret:  os.Getenv("LINE_MESSAGING_API_CHANNEL_SECRET"),
		token:   os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN"),
	}
}

func (c messagingConfig) validate() error {
	if c == (messagingConfig{}) {
		return nil
	}
	for _, item := range []struct{ name, value string }{
		{"LINE_MESSAGING_API_CHANNEL_ID", c.channel},
		{"LINE_MESSAGING_API_CHANNEL_SECRET", c.secret},
		{"LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN", c.token},
	} {
		if strings.TrimSpace(item.value) == "" {
			return fmt.Errorf("missing configuration: %s", item.name)
		}
	}
	if !validChannelID(c.channel) {
		return errors.New("LINE_MESSAGING_API_CHANNEL_ID must contain only digits")
	}
	return nil
}

func validChannelID(id string) bool {
	return id != "" && strings.Trim(id, "0123456789") == ""
}

func lineMembership(token string, timeout time.Duration) *lineplatform.Membership {
	if token == "" {
		return nil
	}
	return &lineplatform.Membership{
		Token: token, Timeout: timeout,
		Client: &http.Client{Transport: httptransport.TraceTransport{}, Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}
}
