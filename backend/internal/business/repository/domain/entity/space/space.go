package space

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInput = errors.New("invalid repository space input")

var ErrNotFound = errors.New("repository space not found")

var ErrCreationConflict = errors.New("repository space creation key conflict")

var ErrUnavailable = errors.New("repository space unavailable")

type Space struct {
	ID, RepositoryID, Name, CreatorID string
	Version                           int64
	CreatedAt, UpdatedAt              time.Time
}

func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 120 {
		return "", ErrInput
	}
	return name, nil
}
