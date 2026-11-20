package parsing

import (
	"errors"

	objectimportDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/objectimport/domain/entity"
)

var ErrNotConfigured = errors.New("object import model not configured")

var ErrModel = errors.New("object import model unavailable")

var ErrOutput = errors.New("object import model output rejected")

type Input struct {
	Source  objectimportDomainEntity.Source
	Consent bool
}

type Result struct {
	Rows       []objectimportDomainEntity.Candidate
	SourceHash string
}
