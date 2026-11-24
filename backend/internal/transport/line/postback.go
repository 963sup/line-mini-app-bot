package line

import "net/url"

// ParsePostback requires one unambiguous action. Each capability validates its own parameters.
func ParsePostback(data string) (url.Values, bool) {
	values, err := url.ParseQuery(data)
	return values, err == nil && len(values["action"]) == 1 && values.Get("action") != ""
}
