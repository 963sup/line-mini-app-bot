package summarization

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var summaryCitationPattern = regexp.MustCompile(`\[S([0-9]+)\]`)

func validateSummary(text string, sourceCount int) (string, error) {
	result := strings.TrimSpace(text)
	if result == "" || utf8.RuneCountInString(result) > 4000 {
		return "", errors.New("invalid summary length")
	}
	citations := summaryCitationPattern.FindAllStringSubmatch(result, -1)
	if len(citations) == 0 {
		return "", errors.New("summary omitted sources")
	}
	for _, citation := range citations {
		n, err := strconv.Atoi(citation[1])
		if err != nil || n < 1 || n > sourceCount {
			return "", errors.New("summary invented a source")
		}
	}
	return result, nil
}
