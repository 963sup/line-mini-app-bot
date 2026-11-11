package lineui

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestQuickMenuPayload(t *testing.T) {
	for _, tc := range []struct {
		group    bool
		commands []string
	}{
		{false, []string{"出勤紀錄"}},
		{true, []string{"打卡", "出勤", "請假", "補卡", "摘要"}},
	} {
		data, err := json.Marshal(QuickMenu(tc.group, "https://liff.line.me/example", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
		if err != nil {
			t.Fatal(err)
		}
		var payload struct {
			Type       string `json:"type"`
			QuickReply struct {
				Items []struct {
					Type   string `json:"type"`
					Action struct {
						Type  string `json:"type"`
						Text  string `json:"text"`
						Label string `json:"label"`
						Uri   string `json:"uri"`
					} `json:"action"`
				} `json:"items"`
			} `json:"quickReply"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte("關閉摘要")) || bytes.Contains(data, []byte("開啟摘要")) {
			t.Fatal("legacy switches must not appear")
		}
		if payload.Type != "text" || len(payload.QuickReply.Items) != len(tc.commands)+2+map[bool]int{true: 1}[tc.group] {
			t.Fatalf("invalid quick reply: %s", data)
		}
		for i, command := range tc.commands {
			index := i
			if tc.group && i > 0 {
				index++
			}
			item := payload.QuickReply.Items[index]
			if item.Type != "action" || item.Action.Type != "message" || item.Action.Text != command || item.Action.Label != command {
				t.Fatalf("invalid command %q: %s", command, data)
			}
		}
		if tc.group {
			journal := payload.QuickReply.Items[1].Action
			if journal.Type != "uri" || journal.Label != "日誌" || journal.Uri != "https://liff.line.me/example?journal=1" {
				t.Fatalf("invalid journal link: %s", data)
			}
		}
		space := payload.QuickReply.Items[len(payload.QuickReply.Items)-2].Action
		expectedSpace := "https://liff.line.me/example?repositories=1"
		if tc.group {
			expectedSpace = "https://liff.line.me/example?repositories=1&repository=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}
		if space.Type != "uri" || space.Label != "空間" || space.Uri != expectedSpace {
			t.Fatalf("invalid repository link: %s", data)
		}
		last := payload.QuickReply.Items[len(payload.QuickReply.Items)-1]
		if last.Action.Type != "uri" || last.Action.Uri != "https://liff.line.me/example" {
			t.Fatalf("invalid URI: %s", data)
		}
	}
}
