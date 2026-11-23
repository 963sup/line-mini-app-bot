package domain

import "errors"

var (
	ErrNotFound = errors.New("enterprise operation is not available")
	ErrConflict = errors.New("organization already belongs to another enterprise")
)
