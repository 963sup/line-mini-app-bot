package attendance

import "errors"

var (
	ErrUnavailable    = errors.New("attendance storage unavailable")
	ErrExpiredLaunch  = errors.New("launch expired")
	ErrInvalidRequest = errors.New("invalid or expired attendance request")
)
