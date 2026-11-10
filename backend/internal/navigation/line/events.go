package line

import (
	"net/url"
	"os"
	"strings"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"

	"github.com/963sup/line-mini-app-bot/internal/navigation/lineui"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func ReplyForEvent(event webhook.EventInterface, appURL string) transport.Reply {
	var source webhook.SourceInterface
	var command, token string
	switch e := event.(type) {
	case webhook.JoinEvent:
		g, ok := e.Source.(webhook.GroupSource)
		if !ok || g.GroupId == "" || !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return transport.Reply{}
		}
		return transport.Reply{Token: e.ReplyToken, Message: lineui.Help(true, appURL)}
	case webhook.MessageEvent:
		if !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return transport.Reply{}
		}
		if sticker, ok := e.Message.(webhook.StickerMessageContent); ok {
			g, group := e.Source.(webhook.GroupSource)
			if !group || g.GroupId == "" || sticker.PackageId == "" || sticker.StickerId == "" ||
				sticker.PackageId != os.Getenv("LINE_MENU_STICKER_PACKAGE_ID") || sticker.StickerId != os.Getenv("LINE_MENU_STICKER_ID") {
				return transport.Reply{}
			}
			return transport.Reply{Token: e.ReplyToken, Message: lineui.QuickMenu(true, appURL)}
		}
		text, ok := e.Message.(webhook.TextMessageContent)
		if !ok {
			return transport.Reply{}
		}
		source, token = e.Source, e.ReplyToken
		command = strings.ToLower(strings.TrimSpace(text.Text))
	case webhook.PostbackEvent:
		if !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) || e.Postback == nil {
			return transport.Reply{}
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || len(values) != 1 || len(values["action"]) != 1 {
			return transport.Reply{}
		}
		switch values.Get("action") {
		case "help":
			command = "/help"
		case "status":
			command = "/status"
		default:
			return transport.Reply{}
		}
		source, token = e.Source, e.ReplyToken
	default:
		return transport.Reply{}
	}

	group := false
	switch s := source.(type) {
	case webhook.UserSource:
		if s.UserId == "" {
			return transport.Reply{}
		}
	case webhook.GroupSource:
		if s.GroupId == "" {
			return transport.Reply{}
		}
		group = true
	default:
		return transport.Reply{}
	}

	switch command {
	case "/menu", "選單", "功能":
		return transport.Reply{Token: token, Message: lineui.QuickMenu(group, appURL)}
	case "/help":
		return transport.Reply{Token: token, Message: lineui.Help(group, appURL)}
	case "/status":
		text := "一對一連線已接通；輸入「出勤紀錄」查詢本人紀錄。"
		if group {
			text = "群組連線已接通；輸入「選單」查看打卡、出勤看板、請假與漏下班檢查。"
		}
		return transport.Reply{Token: token, Message: &messaging_api.TextMessage{Text: text}}
	default:
		// Ordinary group conversation must not trigger the bot or expose personal data.
		if group {
			return transport.Reply{}
		}
		return transport.Reply{Token: token, Message: lineui.Help(false, appURL)}
	}
}
