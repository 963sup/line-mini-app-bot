package entity

import (
	"errors"
)

var ErrForbidden = errors.New("enterprise provisioning not permitted")

var ErrInvalid = errors.New("invalid organization input")

var ErrNotFound = errors.New("organization resource not found")

var ErrConflict = errors.New("conflicting team management")
