package weather

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

const fixture = `{"success":"true","records":{"Locations":[{"Location":[{"LocationName":"臺北市","Latitude":"25.03","Longitude":"121.56","WeatherElement":[{"ElementName":"天氣預報綜合描述","Time":[{"StartTime":"2026-11-08T06:00:00+08:00","EndTime":"2026-11-08T18:00:00+08:00","ElementValue":[{"WeatherDescription":"多雲。降雨機率20%。溫度攝氏22至26度。"}]}]}]},{"LocationName":"臺中市","Latitude":"24.14","Longitude":"120.68","WeatherElement":[{"ElementName":"天氣預報綜合描述","Time":[{"StartTime":"2026-11-08T06:00:00+08:00","EndTime":"2026-11-08T18:00:00+08:00","ElementValue":[{"WeatherDescription":"晴。"}]}]}]}]}]}}`

func TestSelectRegionalForecast(t *testing.T) {
	var data response
	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatal(err)
	}
	now, _ := time.Parse(time.RFC3339, "2026-11-08T12:00:00+08:00")
	for _, tc := range []struct {
		name, address string
		lat, lon      float64
		region        string
		nearby        bool
	}{
		{"address alias", "台北市信義區", 25.03, 121.56, "臺北市", false},
		{"nearest fallback", "", 24.14, 120.68, "臺中市", true},
		{"distant address cannot select region", "臺中市", 25.03, 121.56, "臺北市", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := selectForecast(data, tc.lat, tc.lon, tc.address, now)
			if err != nil || f.Region != tc.region || f.Nearby != tc.nearby || f.Description == "" {
				t.Fatalf("forecast=%+v err=%v", f, err)
			}
		})
	}
	if _, err := selectForecast(data, 35, 139, "", now); !errors.Is(err, ErrOutsideCoverage) {
		t.Fatal("foreign location must not use Taiwan forecast")
	}
	if _, err := selectForecast(data, 25.03, 121.56, "", now.Add(24*time.Hour)); err == nil {
		t.Fatal("expired forecast must not be displayed")
	}
}
