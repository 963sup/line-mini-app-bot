package application

import (
	"context"
	"errors"
	"strings"

	"github.com/963sup/line-mini-app-bot/internal/business/agent/application/port"
)

// QuoteModel proposes an untrusted literal quotation, or nil when none is found.
// The use case must still check that the quotation occurs in the source.
type QuoteModel interface {
	Quote(ctx context.Context, source string) (*string, error)
}

// PageReader must authorize the actor against the current resource before returning text.
type PageReader interface {
	ReadPage(ctx context.Context, actor, repository, revision string, page int) (string, error)
}

type Investigator struct {
	Model QuoteModel
	Pages PageReader
}

var _ port.Investigator = (*Investigator)(nil)

func (s *Investigator) Investigate(ctx context.Context, t port.Investigation) ([]port.Evidence, error) {
	if t.Actor == "" || t.Repository == "" || t.SourceRevision == "" || len(t.AllowedPages) == 0 || len(t.AllowedPages) > 50 || t.StepLimit < 1 {
		return nil, errors.New("invalid investigation scope")
	}
	// One step reads one authorized page and asks for at most one literal quote.
	for i, page := range t.AllowedPages {
		if i >= t.StepLimit {
			break
		}
		if page < 1 {
			return nil, errors.New("invalid source page")
		}
		source, err := s.Pages.ReadPage(ctx, t.Actor, t.Repository, t.SourceRevision, page)
		if err != nil {
			return nil, err
		}
		if len(source) > 32<<10 {
			return nil, errors.New("source page too large")
		}
		quote, err := s.Model.Quote(ctx, source)
		if err != nil {
			return nil, err
		}
		if quote != nil && strings.TrimSpace(*quote) != "" && strings.Contains(source, *quote) {
			return []port.Evidence{{Page: page, Quote: *quote}}, nil
		}
	}
	return nil, nil
}
