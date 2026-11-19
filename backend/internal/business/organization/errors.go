package organization

import "errors"

var (
	ErrInvalid  = errors.New("invalid organization input")
	ErrNotFound = errors.New("organization resource not found")
	ErrConflict = errors.New("conflicting team management")
)
