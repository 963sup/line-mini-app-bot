package attendancehttp

import (
	"regexp"
	"time"
)

type locationInput struct {
	Latitude       *float64   `json:"latitude"`
	Longitude      *float64   `json:"longitude"`
	AccuracyMeters *float64   `json:"accuracyMeters"`
	CapturedAt     *time.Time `json:"capturedAt"`
}

func locationNumberInRange(value, minimum, maximum float64) bool {
	return value >= minimum && value <= maximum
}

var requestIDPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
