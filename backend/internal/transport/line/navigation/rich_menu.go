package navigation

import (
	_ "embed"
	"encoding/json"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

//go:embed rich_menu.json
var attendanceMenuJSON []byte

var attendanceMenu = loadAttendanceMenu()

type menuPage struct {
	Alias, Label, Postback string
}

type menuLayout struct {
	Width, Height, TabHeight, FooterHeight  int64
	TabWidths                               []int64
	ChatBarText, PrivateLabel, MiniAppLabel string
	Pages                                   []menuPage
}

func loadAttendanceMenu() menuLayout {
	var layout menuLayout
	if err := json.Unmarshal(attendanceMenuJSON, &layout); err != nil {
		panic(err)
	}
	if layout.Width != 2500 || layout.Height != 1686 || layout.TabHeight <= 0 || layout.FooterHeight <= 0 ||
		layout.TabHeight+layout.FooterHeight >= layout.Height || layout.PrivateLabel == "" || layout.MiniAppLabel == "" ||
		len(layout.Pages) != 3 || len(layout.TabWidths) != len(layout.Pages) {
		panic("invalid attendance rich menu layout")
	}
	var width int64
	aliases := map[string]bool{}
	for i, page := range layout.Pages {
		if layout.TabWidths[i] <= 0 || page.Alias == "" || aliases[page.Alias] || page.Label == "" || page.Postback == "" {
			panic("invalid attendance rich menu page")
		}
		width += layout.TabWidths[i]
		aliases[page.Alias] = true
	}
	if width != layout.Width {
		panic("invalid attendance rich menu tab widths")
	}
	return layout
}

// AttendanceRichMenuAliases lists the pages in publication order, with the default first.
func AttendanceRichMenuAliases() []string {
	aliases := make([]string, 0, len(attendanceMenu.Pages))
	for _, page := range attendanceMenu.Pages {
		aliases = append(aliases, page.Alias)
	}
	return aliases
}

func AttendanceRichMenus(appURL string) map[string]*messaging_api.RichMenuRequest {
	layout := attendanceMenu
	miniAppURL := featureURL(appURL, "attendance", "")
	if miniAppURL == "" {
		return nil
	}
	menus := make(map[string]*messaging_api.RichMenuRequest, len(layout.Pages))
	for i, page := range layout.Pages {
		menu := &messaging_api.RichMenuRequest{
			Size: &messaging_api.RichMenuSize{Width: layout.Width, Height: layout.Height},
			Name: "出勤：" + page.Label, ChatBarText: layout.ChatBarText,
		}
		var x int64
		for j, target := range layout.Pages {
			width := layout.TabWidths[j]
			if j != i {
				menu.Areas = append(menu.Areas, messaging_api.RichMenuArea{
					Bounds: &messaging_api.RichMenuBounds{X: x, Width: width, Height: layout.TabHeight},
					Action: &messaging_api.RichMenuSwitchAction{Label: target.Label, RichMenuAliasId: target.Alias, Data: "tab=" + target.Alias},
				})
			}
			x += width
		}
		footerY := layout.Height - layout.FooterHeight
		menu.Areas = append(menu.Areas,
			messaging_api.RichMenuArea{
				Bounds: &messaging_api.RichMenuBounds{Y: layout.TabHeight, Width: layout.Width, Height: footerY - layout.TabHeight},
				Action: &messaging_api.PostbackAction{Label: page.Label, Data: page.Postback},
			},
			messaging_api.RichMenuArea{
				Bounds: &messaging_api.RichMenuBounds{Y: footerY, Width: layout.Width / 2, Height: layout.FooterHeight},
				Action: &messaging_api.RichMenuSwitchAction{Label: layout.PrivateLabel, RichMenuAliasId: PrivateRichMenuAlias, Data: "tab=" + PrivateRichMenuAlias},
			},
			messaging_api.RichMenuArea{
				Bounds: &messaging_api.RichMenuBounds{X: layout.Width / 2, Y: footerY, Width: layout.Width / 2, Height: layout.FooterHeight},
				Action: &messaging_api.UriAction{Label: layout.MiniAppLabel, Uri: miniAppURL},
			},
		)
		menus[page.Alias] = menu
	}
	return menus
}

const PrivateRichMenuAlias = "rich-menu-private"

func PrivateRichMenu(appURL string) *messaging_api.RichMenuRequest {
	entries := [...]struct{ label, feature string }{
		{"出勤", "attendance"}, {"補卡", "missing-punch"}, {"更正", "correction"},
		{"請假", "leave"}, {"審核", "review"}, {"儲存庫", "repositories"},
	}
	const width, height int64 = 2500, 1686
	menu := &messaging_api.RichMenuRequest{
		Size: &messaging_api.RichMenuSize{Width: width, Height: height},
		Name: "個人功能", ChatBarText: "開啟選單",
	}
	for i, entry := range entries {
		link := featureURL(appURL, entry.feature, "")
		if link == "" {
			return nil
		}
		column := int64(i % 3)
		x := column * width / 3
		menu.Areas = append(menu.Areas, messaging_api.RichMenuArea{
			Bounds: &messaging_api.RichMenuBounds{X: x, Y: int64(i/3) * height / 2, Width: (column+1)*width/3 - x, Height: height / 2},
			Action: &messaging_api.UriAction{Label: entry.label, Uri: link},
		})
	}
	return menu
}
