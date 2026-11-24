package dto

type TemplateDraft struct {
	Draft                TemplateCreation
	Warnings             []string
	Candidates           []Candidate
	CandidateIndexByNode map[string]int
}
