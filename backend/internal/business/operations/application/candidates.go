package application

import (
	"context"
	"encoding/json"
	"errors"
)

var (
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
	Name     string  `json:"name"`
	Quantity *string `json:"quantity"`
	Unit     *string `json:"unit"`
	Location *string `json:"location"`
	Quote    string  `json:"quote"`
}

// Extraction contains provider diagnostics even when interpreting a response fails.
type Extraction struct {
	Candidates  []Candidate
	RawResponse json.RawMessage
}

type CandidateExtractor interface {
	ExtractCandidates(context.Context, Document) (Extraction, error)
}
