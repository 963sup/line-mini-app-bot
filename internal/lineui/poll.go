package lineui

import (
	"fmt"
	"github.com/963sup/line-mini-app-bot/internal/poll"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"net/url"
	"strconv"
	"strings"
)

func PollMessage(p poll.Poll) messaging_api.MessageInterface {
	state := "開放中（可改票）"
	if p.Closed {
		state = "已結束"
	}
	var text strings.Builder
	fmt.Fprintf(&text, "%s\n%s\n", p.Question, state)
	total := 0
	buttons := []messaging_api.FlexComponentInterface{}
	for i, option := range p.Options {
		count := p.Counts[i]
		total += count
		fmt.Fprintf(&text, "%s：%d 票\n", option, count)
		if !p.Closed {
			data := url.Values{"action": {"poll-vote"}, "poll": {p.ID}, "option": {strconv.Itoa(i)}}.Encode()
			buttons = append(buttons, &messaging_api.FlexButton{Action: &messaging_api.PostbackAction{Label: fmt.Sprintf("選項 %d", i+1), Data: data}})
		}
	}
	fmt.Fprintf(&text, "共 %d 人\n投票結果 %s\n結束投票 %s（建立者限定）", total, p.ID, p.ID)
	bubble := &messaging_api.FlexBubble{Body: &messaging_api.FlexBox{Layout: "vertical", Contents: []messaging_api.FlexComponentInterface{&messaging_api.FlexText{Text: text.String(), Wrap: true}}}}
	if len(buttons) > 0 {
		bubble.Footer = &messaging_api.FlexBox{Layout: "vertical", Contents: buttons}
	}
	return &messaging_api.FlexMessage{AltText: "群組投票：" + p.Question, Contents: bubble}
}
