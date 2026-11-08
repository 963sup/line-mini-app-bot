package main

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/lineui"
	"github.com/963sup/line-mini-app-bot/internal/poll"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pollBot struct{ store poll.Store }

var pollIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func (b *pollBot) handle(ctx context.Context, event webhook.EventInterface) (lineReply, bool, error) {
	var source webhook.SourceInterface
	var token, id, eventID, command string
	option := -1
	switch e := event.(type) {
	case webhook.MessageEvent:
		if e.Mode != "active" {
			return lineReply{}, false, nil
		}
		m, ok := e.Message.(webhook.TextMessageContent)
		if !ok {
			return lineReply{}, false, nil
		}
		text := strings.TrimSpace(m.Text)
		switch {
		case strings.HasPrefix(text, "投票 "):
			command = "create"
		case strings.HasPrefix(text, "投票結果 "):
			command = "get"
			id = strings.TrimSpace(strings.TrimPrefix(text, "投票結果 "))
		case strings.HasPrefix(text, "結束投票 "):
			command = "close"
			id = strings.TrimSpace(strings.TrimPrefix(text, "結束投票 "))
		default:
			return lineReply{}, false, nil
		}
		source, token, eventID = e.Source, e.ReplyToken, e.WebhookEventId
		if command == "create" {
			id = text
		}
	case webhook.PostbackEvent:
		if e.Mode != "active" || e.Postback == nil {
			return lineReply{}, false, nil
		}
		values, err := url.ParseQuery(e.Postback.Data)
		if err != nil || values.Get("action") != "poll-vote" {
			return lineReply{}, false, nil
		}
		source, token, eventID = e.Source, e.ReplyToken, e.WebhookEventId
		command, id = "vote", values.Get("poll")
		option, err = strconv.Atoi(values.Get("option"))
		if err != nil || len(values) != 3 || len(values["action"]) != 1 || len(values["poll"]) != 1 || len(values["option"]) != 1 {
			option = -1
		}
	default:
		return lineReply{}, false, nil
	}
	respond := func(text string) (lineReply, bool, error) {
		return lineReply{token: token, message: &messaging_api.TextMessage{Text: text}}, true, nil
	}
	group, ok := source.(webhook.GroupSource)
	if !ok || group.GroupId == "" || group.UserId == "" {
		return respond("請在群組中使用投票，且須能辨識投票者。")
	}
	if !pollIDPattern.MatchString(eventID) || (command != "create" && !pollIDPattern.MatchString(id)) {
		return respond("投票 ID 或事件資料無效。")
	}
	if b == nil || b.store == nil {
		return respond("投票尚未啟用。")
	}
	var p poll.Poll
	var err error
	switch command {
	case "create":
		p, err = poll.Parse(id)
		if err != nil {
			return respond("格式：投票 題目｜選項A｜選項B（2–4 個不重複選項；題目最多100字、選項40字）")
		}
		var result poll.Result
		result, err = b.store.Create(ctx, group.GroupId, group.UserId, eventID, p)
		p = result.Poll
	case "vote":
		if option < 0 {
			return respond("投票選項無效。")
		}
		var result poll.Result
		result, err = b.store.Vote(ctx, group.GroupId, group.UserId, eventID, id, option)
		p = result.Poll
	case "get":
		p, err = b.store.Get(ctx, group.GroupId, id)
	case "close":
		p, err = b.store.Close(ctx, group.GroupId, group.UserId, id)
	}
	if errors.Is(err, poll.ErrInvalid) || status.Code(err) == codes.NotFound {
		return respond("投票不存在、已結束，或您無權執行此操作。")
	}
	if err != nil {
		return lineReply{}, true, err
	}
	return lineReply{token: token, message: lineui.PollMessage(p)}, true, nil
}
