package navigation

import (
	"context"
	"net/url"
	"os"
	"strings"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type menuRequest struct {
	token   string
	command string
	groupID string
}

func parseMenu(event webhook.EventInterface) menuRequest {
	var source webhook.SourceInterface
	var command, token string
	switch e := event.(type) {
	case webhook.JoinEvent:
		g, ok := e.Source.(webhook.GroupSource)
		if !ok || g.GroupId == "" || !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return menuRequest{}
		}
		return menuRequest{token: e.ReplyToken, command: "/help", groupID: g.GroupId}
	case webhook.MessageEvent:
		if !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return menuRequest{}
		}
		if sticker, ok := e.Message.(webhook.StickerMessageContent); ok {
			g, group := e.Source.(webhook.GroupSource)
			if !group || g.GroupId == "" || sticker.PackageId == "" || sticker.StickerId == "" ||
				sticker.PackageId != os.Getenv("LINE_MENU_STICKER_PACKAGE_ID") || sticker.StickerId != os.Getenv("LINE_MENU_STICKER_ID") {
				return menuRequest{}
			}
			return menuRequest{token: e.ReplyToken, command: "/menu", groupID: g.GroupId}
		}
		text, ok := e.Message.(webhook.TextMessageContent)
		if !ok {
			return menuRequest{}
		}
		source, token = e.Source, e.ReplyToken
		command = strings.ToLower(strings.TrimSpace(text.Text))
	case webhook.PostbackEvent:
		if !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) || e.Postback == nil {
			return menuRequest{}
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || len(values) != 1 || len(values["action"]) != 1 {
			return menuRequest{}
		}
		if values.Get("action") != "help" && values.Get("action") != "miniapp" {
			return menuRequest{}
		}
		command = "/" + values.Get("action")
		source, token = e.Source, e.ReplyToken
	default:
		return menuRequest{}
	}

	var groupID string
	switch s := source.(type) {
	case webhook.UserSource:
		if s.UserId == "" {
			return menuRequest{}
		}
	case webhook.GroupSource:
		if s.GroupId == "" {
			return menuRequest{}
		}
		groupID = s.GroupId
	default:
		return menuRequest{}
	}

	switch command {
	case "選單", "功能":
		command = "/menu"
	case "/menu", "/miniapp", "/help":
	default:
		if groupID != "" {
			return menuRequest{}
		}
		command = "private-help"
	}
	return menuRequest{token: token, command: command, groupID: groupID}
}

func (m menuRequest) reply(appURL, endpointURL string, ids ...string) transport.Reply {
	reply := transport.Reply{Token: m.token}
	switch m.command {
	case "/menu":
		reply.Message = QuickMenu(m.groupID != "", appURL, ids...)
	case "/miniapp":
		reply.Message = MiniApp(appURL, endpointURL, ids...)
	case "/help":
		reply.Message = Help(m.groupID != "", appURL, endpointURL, ids...)
	case "private-help":
		reply.Message = Help(false, appURL, endpointURL)
	}
	return reply
}

func ReplyForEvent(event webhook.EventInterface, appURL, endpointURL string, ids ...string) transport.Reply {
	return parseMenu(event).reply(appURL, endpointURL, ids...)
}

func Handler(appURL, endpointURL string, lookup func(context.Context, string) (string, error)) transport.Handler {
	return transport.Handler{Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
		menu := parseMenu(event)
		if menu.groupID != "" && lookup != nil {
			id, err := lookup(ctx, menu.groupID)
			if err != nil {
				return transport.Reply{}, false, err
			}
			return menu.reply(appURL, endpointURL, id), true, nil
		}
		return menu.reply(appURL, endpointURL), true, nil
	}}
}
