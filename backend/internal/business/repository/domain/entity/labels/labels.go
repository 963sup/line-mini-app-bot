package labels

import (
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

var ErrInput = errors.New("invalid label input")

var ErrNotFound = errors.New("repository not found")

var ErrStaleName = errors.New("label name has changed")

var ErrNameConflict = errors.New("label name already exists")

var ErrCreationConflict = errors.New("label creation key conflict")

var ErrUnavailable = errors.New("label storage unavailable")

type Label struct {
	ID, RepositoryID, Name string
	CreatedAt              time.Time
}

func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 40 {
		return "", ErrInput
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", ErrInput
		}
	}
	return name, nil
}

func NameKey(name string) string { return strings.ToLower(name) }

// NormalizeIDs treats references as a bounded set and returns an independent slice.
func NormalizeIDs(ids []string) ([]string, error) {
	if len(ids) > 10 {
		return nil, ErrInput
	}
	result := append([]string{}, ids...)
	sort.Strings(result)
	for i, id := range result {
		if !repositoryDomainEntity.ValidID(id) || (i > 0 && id == result[i-1]) {
			return nil, ErrInput
		}
	}
	return result, nil
}
