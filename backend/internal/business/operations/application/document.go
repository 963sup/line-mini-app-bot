package application

import (
	"context"
	"errors"
)

// DocumentTextExtractor supplies UTF-8 text, preserving page breaks.
// Implementations do not persist documents or extracted text.
type DocumentTextExtractor interface {
	ExtractText(context.Context, Document) (string, error)
}

var ErrUnauthenticated = errors.New("document extraction requires identity")

// DocumentTextService provides a read-only tool; documents and text are not saved.
type DocumentTextService struct{ Extractor DocumentTextExtractor }

func (s *DocumentTextService) ExtractText(ctx context.Context, actor string, doc Document) (string, error) {
	if actor == "" {
		return "", ErrUnauthenticated
	}
	if s == nil || s.Extractor == nil {
		return "", ErrProvider
	}
	return s.Extractor.ExtractText(ctx, doc)
}
