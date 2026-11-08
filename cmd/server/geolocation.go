package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"math"
	"mime"
	"net/http"
	"time"
)

//go:embed web/geolocation.html
var geolocationPage []byte

//go:embed web/geolocation.js
var geolocationScript []byte

//go:embed web/geolocation.css
var geolocationStyle []byte

type locationInput struct {
	Latitude       *float64   `json:"latitude"`
	Longitude      *float64   `json:"longitude"`
	AccuracyMeters *float64   `json:"accuracyMeters"`
	CapturedAt     *time.Time `json:"capturedAt"`
}

func registerGeolocation(mux *http.ServeMux) {
	mux.HandleFunc("GET /location", staticAssetHandler(geolocationPage, "text/html; charset=utf-8"))
	mux.HandleFunc("GET /assets/geolocation.js", staticAssetHandler(geolocationScript, "text/javascript; charset=utf-8"))
	mux.HandleFunc("GET /assets/geolocation.css", staticAssetHandler(geolocationStyle, "text/css; charset=utf-8"))
	protection := http.NewCrossOriginProtection()
	mux.Handle("POST /api/location/validate", protection.Handler(http.HandlerFunc(validateLocation)))
}

// validateLocation is a public, stateless diagnostic. It grants no attendance or identity claims.
func validateLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		locationError(w, http.StatusUnsupportedMediaType, "請使用 application/json。")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input locationInput
	if err := decoder.Decode(&input); err != nil {
		locationDecodeError(w, err)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		locationDecodeError(w, err)
		return
	}
	if input.Latitude == nil || input.Longitude == nil {
		locationError(w, http.StatusBadRequest, "缺少經緯度。")
		return
	}
	if !locationNumberInRange(*input.Latitude, -90, 90) || !locationNumberInRange(*input.Longitude, -180, 180) {
		locationError(w, http.StatusBadRequest, "經緯度超出有效範圍。")
		return
	}
	if input.AccuracyMeters == nil || !locationNumberInRange(*input.AccuracyMeters, 0, math.MaxFloat64) {
		locationError(w, http.StatusBadRequest, "缺少有效的位置精度。")
		return
	}
	now := time.Now().UTC()
	if input.CapturedAt == nil || input.CapturedAt.IsZero() {
		locationError(w, http.StatusBadRequest, "缺少有效的採樣時間。")
		return
	}
	if input.CapturedAt.Before(now.Add(-5*time.Minute)) || input.CapturedAt.After(now.Add(time.Minute)) {
		locationError(w, http.StatusBadRequest, "位置時間過舊或超前，請確認裝置時間並重新定位。")
		return
	}
	// Do not log, persist, or echo the submitted coordinates.
	_ = json.NewEncoder(w).Encode(struct {
		Valid      bool      `json:"valid"`
		ReceivedAt time.Time `json:"receivedAt"`
	}{Valid: true, ReceivedAt: now})
}

func locationNumberInRange(value, minimum, maximum float64) bool {
	return value >= minimum && value <= maximum
}

func locationDecodeError(w http.ResponseWriter, err error) {
	var sizeError *http.MaxBytesError
	if errors.As(err, &sizeError) {
		locationError(w, http.StatusRequestEntityTooLarge, "定位資料過大。")
		return
	}
	locationError(w, http.StatusBadRequest, "定位資料格式錯誤，請重新取得位置。")
}

func locationError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: message})
}
