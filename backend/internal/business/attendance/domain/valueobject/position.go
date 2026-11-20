package valueobject

import (
	"time"
)

type Position struct {
	Latitude       float64
	Longitude      float64
	AccuracyMeters float64
	CapturedAt     time.Time
}

// ValidAt checks coordinates and freshness without changing captured evidence.
func (v Position) ValidAt(now time.Time) bool {
	return v.Latitude >= -90 && v.Latitude <= 90 && v.Longitude >= -180 && v.Longitude <= 180 &&
		v.AccuracyMeters >= 0 && v.AccuracyMeters <= 1.7976931348623157e308 &&
		!v.CapturedAt.Before(now.Add(-5*time.Minute)) && !v.CapturedAt.After(now.Add(time.Minute))
}
