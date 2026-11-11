// Package weather queries CWA regional forecasts without retaining user locations.
package cwa

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/weather/forecast"
)

type Client struct {
	APIKey string
	HTTP   *http.Client
}

type period struct {
	StartTime, EndTime string
	ElementValue       []struct{ WeatherDescription string }
}
type location struct {
	LocationName, Latitude, Longitude string
	WeatherElement                    []struct {
		ElementName string
		Time        []period
	}
}
type response struct {
	Success string
	Records struct {
		Locations []struct{ Location []location }
	}
}

func (c *Client) Lookup(ctx context.Context, lat, lon float64, address string) (forecast.Forecast, error) {
	if !(lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180) {
		return forecast.Forecast{}, forecast.ErrOutsideCoverage
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	q := url.Values{"Authorization": {c.APIKey}, "format": {"JSON"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://opendata.cwa.gov.tw/api/v1/rest/datastore/F-D0047-091?"+q.Encode(), nil)
	if err != nil {
		return forecast.Forecast{}, errors.New("cannot create forecast request")
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return forecast.Forecast{}, errors.New("forecast service unavailable")
	}
	// The response is read-only; closing cannot commit application state.
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return forecast.Forecast{}, errors.New("forecast service rejected request")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, (4<<20)+1))
	if err != nil || len(body) > 4<<20 {
		return forecast.Forecast{}, errors.New("invalid forecast response size")
	}
	var data response
	if err := json.Unmarshal(body, &data); err != nil || data.Success != "true" {
		return forecast.Forecast{}, errors.New("invalid forecast response")
	}
	return selectForecast(data, lat, lon, address, time.Now())
}

func selectForecast(data response, lat, lon float64, address string, now time.Time) (forecast.Forecast, error) {
	address = strings.ReplaceAll(address, "台", "臺")
	var nearest, matched *location
	best := math.Inf(1)
	for _, area := range data.Records.Locations {
		for i := range area.Location {
			loc := &area.Location[i]
			x, e1 := strconv.ParseFloat(loc.Latitude, 64)
			y, e2 := strconv.ParseFloat(loc.Longitude, 64)
			if e1 != nil || e2 != nil || math.IsNaN(x) || math.IsNaN(y) || math.Abs(x) > 90 || math.Abs(y) > 180 {
				continue
			}
			d := distance(lat, lon, x, y)
			if d < best {
				nearest = loc
				best = d
			}
			if loc.LocationName != "" && strings.Contains(address, strings.ReplaceAll(loc.LocationName, "台", "臺")) && d <= 100 {
				matched = loc
			}
		}
	}
	chosen := matched
	if chosen == nil {
		if nearest == nil || best > 60 {
			return forecast.Forecast{}, forecast.ErrOutsideCoverage
		}
		chosen = nearest
	}
	for _, element := range chosen.WeatherElement {
		if element.ElementName != "天氣預報綜合描述" {
			continue
		}
		for _, p := range element.Time {
			start, e1 := time.Parse(time.RFC3339, p.StartTime)
			end, e2 := time.Parse(time.RFC3339, p.EndTime)
			if e1 != nil || e2 != nil || now.Before(start) || !now.Before(end) {
				continue
			}
			for _, value := range p.ElementValue {
				if strings.TrimSpace(value.WeatherDescription) != "" {
					return forecast.Forecast{Region: chosen.LocationName, Description: value.WeatherDescription, Start: start, End: end, Nearby: matched == nil}, nil
				}
			}
		}
	}
	return forecast.Forecast{}, errors.New("no current forecast available")
}

func distance(lat, lon, x, y float64) float64 {
	const radians = math.Pi / 180
	a := math.Pow(math.Sin((x-lat)*radians/2), 2) + math.Cos(lat*radians)*math.Cos(x*radians)*math.Pow(math.Sin((y-lon)*radians/2), 2)
	return 6371 * 2 * math.Asin(math.Sqrt(math.Min(1, a)))
}
