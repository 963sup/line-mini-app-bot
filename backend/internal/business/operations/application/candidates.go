package application

import (
	"context"
	"errors"
)

var (
	ErrProvider        = errors.New("model request failed")
	ErrRefused         = errors.New("candidate extraction refused")
	ErrIncomplete      = errors.New("candidate extraction incomplete")
	ErrInvalidOutput   = errors.New("invalid candidate response")
	ErrInvalidDocument = errors.New("invalid document")
)

type Document struct {
	Filename string
	Data     []byte
}

type Candidate struct {
	CategoryPath []string
	Name         string
	Quantity     *string
	Unit         *string
	Location     *string
	Quote        string
}

// Extraction contains provider diagnostics even when interpreting a response fails.
type Extraction struct {
	Candidates  []Candidate
	RawResponse []byte
}

type CandidateExtractor interface {
	ExtractCandidates(context.Context, Document) (Extraction, error)
}
