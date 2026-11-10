package weather

import (
	"errors"
	"time"
)

var ErrOutsideCoverage = errors.New("location outside forecast coverage")

type Forecast struct {
	Region, Description string
	Start, End          time.Time
	Nearby              bool
}
