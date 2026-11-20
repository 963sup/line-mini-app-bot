package port

import (
	"errors"
)

var ErrUnavailable = errors.New("attendance storage unavailable")

var ErrExpiredLaunch = errors.New("launch expired")

var ErrInvalidRequest = errors.New("invalid or expired attendance request")
