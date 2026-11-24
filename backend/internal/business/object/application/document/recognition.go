package document

import (
	"context"
	"strings"
)

// recognizeDocumentObjects extracts candidates and locates missing page references.
// A nil location leaves the evidence for manual review.
func recognizeDocumentObjects(
	ctx context.Context,
	extract CandidateExtractor,
	input Document,
	readText func(context.Context) (string, error),
) (Extraction, error) {
	result, err := extract.ExtractCandidates(ctx, input)
	if err != nil || readText == nil {
		return result, err
	}
	var find func(string) []int
	const maxLookups = 32
	lookups := 0
	items := result.Candidates
	for i := range items {
		if items[i].Location != nil && strings.TrimSpace(*items[i].Location) != "" {
			continue
		}
		items[i].Location = nil
		quote := strings.TrimSpace(items[i].Quote)
		if quote == "" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if find == nil {
			source, err := readText(ctx)
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			if err != nil || source == "" {
				return result, nil
			}
			find = IndexPages(source)
		}
		items[i].Location = InferLocation(find(quote))
		lookups++
		if lookups == maxLookups {
			break
		}
	}
	return result, nil
}
