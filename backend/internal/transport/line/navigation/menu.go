// Package navigation renders the public feature menu shared by the bot and CLI.
package navigation

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func JournalURL(appURL string, repositoryIDs ...string) string {
	if appURL == "" {
		return ""
	}
	u, err := url.Parse(appURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	q := u.Query()
	q.Set("journal", "1")
	q.Set("journalCompose", "1")
	if len(repositoryIDs) == 1 && repositoryIDPattern.MatchString(repositoryIDs[0]) {
		q.Set("journalRepository", repositoryIDs[0])
	}
	u.RawQuery = q.Encode()
	return u.String()
}

var repositoryIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func RepositoryURL(appURL string, group bool, ids ...string) string {
	if group && (len(ids) != 1 || !repositoryIDPattern.MatchString(ids[0])) {
		return ""
	}
	if appURL == "" {
		return ""
	}
	u, err := url.Parse(appURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	q := u.Query()
	q.Set("repositories", "1")
	if group {
		q.Set("repository", ids[0])
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// miniAppAction uses only the configured endpoint, never a request host or postback URL.
func miniAppAction(appURL, desktopURL string) *messaging_api.UriAction {
	action := &messaging_api.UriAction{Label: "開啟 MINI App", Uri: appURL}
	endpoint, err := url.Parse(desktopURL)
	if err == nil && len(endpoint.String()) <= 10000 && endpoint.Scheme == "https" && endpoint.Hostname() != "" && endpoint.User == nil && !strings.EqualFold(endpoint.Hostname(), "miniapp.line.me") && !strings.EqualFold(endpoint.Hostname(), "liff.line.me") {
		action.AltUri = &messaging_api.AltUri{Desktop: endpoint.String()}
	}
	return action
}

// MiniApp opens one configured form or resource through the native buttons template.
func MiniApp(appURL, desktopURL string, repositoryIDs ...string) messaging_api.MessageInterface {
	action := miniAppAction(appURL, desktopURL)
	if link := RepositoryURL(appURL, true, repositoryIDs...); link != "" {
		action.Uri = link
		if action.AltUri != nil {
			if desktop := RepositoryURL(action.AltUri.Desktop, true, repositoryIDs...); desktop != "" {
				action.AltUri.Desktop = desktop
			}
		}
	}
	text := "手機開啟 LINE MINI App；電腦開啟網頁版，並依畫面登入 LINE。"
	if action.AltUri == nil {
		text = "此入口目前供手機 LINE 使用；電腦版入口尚未設定。"
	}
	if appURL == "" {
		return &messaging_api.TextMessage{Text: "MINI App 尚未設定。"}
	}
	return &messaging_api.TemplateMessage{AltText: "開啟 MINI App", Template: &messaging_api.ButtonsTemplate{
		Text: text, Actions: []messaging_api.ActionInterface{action},
	}}
}
