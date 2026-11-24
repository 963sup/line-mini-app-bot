package line

// MiniAppURL returns the permanent LINE entry point for a configured LIFF app.
func MiniAppURL(liffID string) string {
	if liffID == "" {
		return ""
	}
	return "https://miniapp.line.me/" + liffID
}
