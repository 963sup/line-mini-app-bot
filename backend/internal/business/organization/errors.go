package organization

import "errors"

var (
	ErrForbidden = errors.New("enterprise provisioning not permitted")
	ErrInvalid   = errors.New("invalid organization input")
	ErrNotFound  = errors.New("organization resource not found")
	ErrConflict  = errors.New("conflicting team management")
)
