// Package navigation renders the public feature menu shared by the bot and CLI.
package navigation

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func journalURL(appURL string) string {
	if appURL == "" {
		return ""
	}
	u, err := url.Parse(appURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	q := u.Query()
	q.Set("journal", "1")
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

func Help(group bool, appURL, desktopURL string, repositoryIDs ...string) messaging_api.MessageInterface {
	title := "LINE 功能選單"
	description := "輸入「出勤紀錄」查詢本人紀錄，或輸入「建立組織」「建立企業」開始建立。"
	if group {
		title = "群組功能選單"
		description = "點選下方按鈕：打卡、出勤、請假登記或補卡申請。請假送出不代表核准；歷史出勤請私訊查詢。分享位置可查天氣，輸入「摘要」整理群聊。"
	}
	buttons := []messaging_api.FlexComponentInterface{}
	if group {
		for _, entry := range []struct{ label, command string }{
			{"打卡", "打卡"}, {"出勤", "出勤"}, {"請假登記", "請假"}, {"補卡申請", "補卡"},
		} {
			buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.MessageAction{Label: entry.label, Text: entry.command}})
			if entry.command == "打卡" && journalURL(appURL) != "" {
				buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.UriAction{Label: "日誌", Uri: journalURL(appURL)}})
			}
		}
	} else {
		buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: "出勤紀錄", Data: "action=attendance-history"}})
	}
	if link := RepositoryURL(appURL, group, repositoryIDs...); link != "" {
		buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.UriAction{Label: "空間", Uri: link}})
	}
	if appURL != "" {
		buttons = append(buttons, &messaging_api.FlexButton{
			Style:  "primary",
			Action: miniAppAction(appURL, desktopURL),
		})
	}
	return &messaging_api.FlexMessage{
		AltText: title,
		Contents: &messaging_api.FlexBubble{
			Body: &messaging_api.FlexBox{
				Layout: "vertical",
				Contents: []messaging_api.FlexComponentInterface{
					&messaging_api.FlexText{Text: title, Weight: "bold", Size: "xl", Wrap: true},
					&messaging_api.FlexText{Text: description, Wrap: true, Margin: "md"},
				},
			},
			Footer: &messaging_api.FlexBox{Layout: "vertical", Contents: buttons},
		},
	}
}

// QuickMenu puts each supported command in a separate reply chip.
func QuickMenu(group bool, appURL string, repositoryIDs ...string) messaging_api.MessageInterface {
	items := []messaging_api.QuickReplyItem{}
	add := func(label, command string) {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.MessageAction{Label: label, Text: command}})
	}
	if group {
		add("打卡", "打卡")
		if link := journalURL(appURL); link != "" {
			items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "日誌", Uri: link}})
		}
		add("出勤", "出勤")
		add("請假", "請假")
		add("補卡", "補卡")
		add("摘要", "摘要")
	} else {
		add("出勤紀錄", "出勤紀錄")
		add("建立組織", "建立組織")
		add("建立企業", "建立企業")
	}
	if link := RepositoryURL(appURL, group, repositoryIDs...); link != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.UriAction{Label: "空間", Uri: link}})
	}
	if appURL != "" {
		items = append(items, messaging_api.QuickReplyItem{Type: "action", Action: &messaging_api.PostbackAction{Label: "MINI App", Data: "action=miniapp"}})
	}
	return &messaging_api.TextMessage{Text: "請選擇功能：", QuickReply: &messaging_api.QuickReply{Items: items}}
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

// MiniApp opens the same mobile MINI App through a Flex-supported desktop alternate.
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
	return &messaging_api.FlexMessage{AltText: "開啟 MINI App", Contents: &messaging_api.FlexBubble{
		Body:   &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{&messaging_api.FlexText{Text: text, Wrap: true}}},
		Footer: &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{&messaging_api.FlexButton{Style: "primary", Action: action}}},
	}}
}
