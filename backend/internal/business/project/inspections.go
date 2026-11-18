package project

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInspectionPending = errors.New("inspection pending")
	ErrInspectionHistory = errors.New("inspection history retained")
)

type InspectionSubmission struct {
	ID                 string
	Note               string
	Status             string
	SubmittedBy        string
	SubmittedAt        time.Time
	SourceType         string
	SourceID           string
	SourceRepositoryID string
	SourceVersion      int64
}

func (s *Service) Inspections(ctx context.Context, actor, ownerID, projectID, itemID string, note *string) ([]InspectionSubmission, error) {
	if !ValidID(itemID) {
		return nil, ErrInput
	}
	var text string
	if note != nil {
		text = strings.TrimSpace(*note)
		if text == "" || !utf8.ValidString(text) || utf8.RuneCountInString(text) > 2000 {
			return nil, ErrInput
		}
	}
	scope, _, err := s.projectScope(ctx, actor, ownerID, projectID)
	if err != nil {
		return nil, err
	}
	return s.Store.Inspections(ctx, scope, projectID, itemID, text)
}
