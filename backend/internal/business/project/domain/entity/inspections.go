package entity

import (
	"errors"
	"time"
)

var ErrInspectionPending = errors.New("inspection pending")

var ErrInspectionHistory = errors.New("inspection history retained")

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
