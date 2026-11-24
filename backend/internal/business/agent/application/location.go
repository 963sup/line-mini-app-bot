package application

import "strconv"

// InferLocation proposes a location only when the source identifies one page.
func InferLocation(pages []int) *string {
	if len(pages) != 1 || pages[0] < 1 {
		return nil
	}
	location := "第 " + strconv.Itoa(pages[0]) + " 頁"
	return &location
}
