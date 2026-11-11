package summary

import "errors"

var ErrUnavailable = errors.New("summary storage unavailable")

var ErrDeliveryUnavailable = errors.New("summary delivery unavailable")
