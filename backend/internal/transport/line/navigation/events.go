package navigation

import (
	"context"
	"maps"
	"slices"
	"strings"

	transport "github.com/963sup/line-mini-app-bot/internal/transport/line"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type menuRequest struct {
	token     string
	command   string
	groupID   string
	textReply string
}

type Menu struct {
	StickerPackageID, StickerID string
	PrivateActions              []PrivateAction
}

func (trigger Menu) routes() routes {
	entries := slices.Clone(commonRoutes)
	for _, a := range trigger.PrivateActions {
		entries = append(entries, route{command: "/" + a.Action, action: a.Action, privatePostback: true, textReply: a.Reply})
	}
	return indexRoutes(entries...)
}

func (trigger Menu) parseMenu(event webhook.EventInterface, menuRoutes routes) menuRequest {
	var source webhook.SourceInterface
	var command, token string
	switch e := event.(type) {
	case webhook.MessageEvent:
		if !transport.CanReply(e.Mode, e.ReplyToken, e.DeliveryContext) {
			return menuRequest{}
		}
		if sticker, ok := e.Message.(webhook.StickerMessageContent); ok {
			g, group := e.Source.(webhook.GroupSource)
			if !group || g.GroupId == "" || sticker.PackageId == "" || sticker.StickerId == "" ||
				sticker.PackageId != trigger.StickerPackageID || sticker.StickerId != trigger.StickerID {
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
		values, ok := transport.ParsePostback(e.Postback.Data)
		if !ok || len(values) != 1 {
			return menuRequest{}
		}
		entry, exists := menuRoutes.postback[values.Get("action")]
		if !exists {
			return menuRequest{}
		}
		if entry.privatePostback {
			if _, ok := e.Source.(webhook.UserSource); !ok {
				return menuRequest{}
			}
		}
		command = entry.command
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

	entry, ok := menuRoutes.text[command]
	if !ok || (groupID != "" && entry.privatePostback) {
		return menuRequest{}
	}
	return menuRequest{token: token, command: entry.command, groupID: groupID, textReply: entry.textReply}
}

func (m menuRequest) reply(appURL, endpointURL string, ids ...string) transport.Reply {
	reply := transport.Reply{Token: m.token}
	switch m.command {
	case "/menu", "/help":
		if m.groupID == "" {
			reply.Message = &messaging_api.TextMessage{Text: "請使用聊天室下方的選單開啟所需功能。"}
		} else {
			reply.Message = groupMenu(appURL, m.groupID, ids...)
		}
	case "/miniapp":
		reply.Message = MiniApp(appURL, endpointURL, ids...)
	default:
		if m.textReply != "" {
			reply.Message = &messaging_api.TextMessage{Text: m.textReply}
		}
	}
	return reply
}

func (trigger Menu) Handler(appURL, endpointURL string, lookup func(context.Context, string) (string, error)) transport.Handler {
	menuRoutes := trigger.routes()
	return transport.Handler{Owner: "navigation", Commands: slices.Sorted(maps.Keys(menuRoutes.text)), Postbacks: slices.Sorted(maps.Keys(menuRoutes.postback)), Handle: func(ctx context.Context, event webhook.EventInterface) (transport.Reply, bool, error) {
		menu := trigger.parseMenu(event, menuRoutes)
		if menu.token == "" {
			return transport.Reply{}, false, nil
		}
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
