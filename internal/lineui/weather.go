package lineui

import (
	"fmt"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/weather"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func Weather(f weather.Forecast) messaging_api.MessageInterface {
	label := "縣市預報"
	if f.Nearby {
		label = "鄰近縣市預報"
	}
	zone := time.FixedZone("Asia/Taipei", 8*60*60)
	return &messaging_api.TextMessage{Text: fmt.Sprintf("%s｜%s\n%s–%s\n\n%s\n\n資料來源：中央氣象署（區域預報，非即時觀測）", f.Region, label, f.Start.In(zone).Format("01/02 15:04"), f.End.In(zone).Format("01/02 15:04"), f.Description)}
}
