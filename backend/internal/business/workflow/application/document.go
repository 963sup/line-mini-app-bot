package application

import "context"

// PrepareDocument coordinates a read-only request. Source text is only needed
// when a candidate has a quotation that still needs locating.
func PrepareDocument[T any, C any](
	ctx context.Context,
	extract func(context.Context) (T, error),
	candidates func(T) []C,
	quoteToLocate func(C) string,
	readText func(context.Context) (string, error),
	index func(string) func(string) []int,
	review func(*C, []int),
) (T, error) {
	result, err := extract(ctx)
	if err != nil || readText == nil {
		return result, err
	}
	var find func(string) []int
	const maxLookups = 32
	lookups := 0
	items := candidates(result)
	for i := range items {
		quote := quoteToLocate(items[i])
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
			find = index(source)
		}
		review(&items[i], find(quote))
		lookups++
		if lookups == maxLookups {
			break
		}
	}
	return result, nil
}
