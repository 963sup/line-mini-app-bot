package application

import "strings"

// IndexPages searches this request's source text. Two matches are enough to
// establish ambiguity. Whitespace normalization handles PDF line wrapping.
func IndexPages(text string) func(string) []int {
	pages := strings.Split(text, "\f")
	for i := range pages {
		pages[i] = normalize(pages[i])
	}
	return func(quote string) []int {
		quote = normalize(quote)
		if quote == "" {
			return nil
		}
		var matches []int
		for i, page := range pages {
			if strings.Contains(page, quote) {
				matches = append(matches, i+1)
				if len(matches) == 2 {
					break
				}
			}
		}
		return matches
	}
}

func normalize(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
