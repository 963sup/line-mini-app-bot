package dto

import (
	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

type TemplateDraft struct {
	Draft                TemplateCreation
	Warnings             []string
	Candidates           []operations.Candidate
	CandidateIndexByNode map[string]int
}
