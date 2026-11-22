package navigation

import (
	_ "embed"
	"encoding/json"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

//go:embed attendance-rich-menu.json
var attendanceMenuJSON []byte

var attendanceMenu = loadAttendanceMenu()

type menuPage struct {
	Alias, Label, Postback string
}

type menuLayout struct {
	Width, Height, TabHeight int64
	TabWidths                []int64
	ChatBarText              string
	Pages                    []menuPage
}

func loadAttendanceMenu() menuLayout {
	var layout menuLayout
	if err := json.Unmarshal(attendanceMenuJSON, &layout); err != nil {
		panic(err)
	}
	if layout.Width != 2500 || layout.Height != 1686 || layout.TabHeight <= 0 || layout.TabHeight >= layout.Height ||
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

func AttendanceRichMenus() map[string]*messaging_api.RichMenuRequest {
	layout := attendanceMenu
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
		menu.Areas = append(menu.Areas, messaging_api.RichMenuArea{
			Bounds: &messaging_api.RichMenuBounds{Y: layout.TabHeight, Width: layout.Width, Height: layout.Height - layout.TabHeight},
			Action: &messaging_api.PostbackAction{Label: page.Label, Data: page.Postback},
		})
		menus[page.Alias] = menu
	}
	return menus
}
