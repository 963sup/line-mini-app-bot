package attendancehttp

import "net/http"

type memberTransport func(*http.Request) (*http.Response, error)

func (f memberTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
