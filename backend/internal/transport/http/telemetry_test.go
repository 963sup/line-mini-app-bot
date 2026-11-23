package httptransport

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTelemetryPropagationMetricsAndPrivacy(t *testing.T) {
	var output bytes.Buffer
	oldTrace, oldMeter := otel.GetTracerProvider(), otel.GetMeterProvider()
	shutdown := NewTelemetry(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { otel.SetTracerProvider(oldTrace); otel.SetMeterProvider(oldMeter) })
	const traceID = "1234567890abcdef1234567890abcdef"
	const private = "private-token-body-and-id"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		request, err := http.NewRequestWithContext(r.Context(), "GET", "https://example.test/?token="+private, nil)
		if err != nil {
			t.Fatal(err)
		}
		transport := TraceTransport{Base: roundTripFunc(func(sent *http.Request) (*http.Response, error) {
			if !strings.Contains(sent.Header.Get("traceparent"), traceID) {
				t.Error("outgoing trace lost its parent")
			}
			return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
		})}
		response, err := transport.RoundTrip(request)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if request.Header.Get("traceparent") != "" {
			t.Error("transport mutated the original request")
		}
		w.WriteHeader(503)
	})
	request := httptest.NewRequest("GET", "/items/"+private+"?token="+private, nil)
	request.Header.Set("traceparent", "00-"+traceID+"-1234567890abcdef-01")
	w := httptest.NewRecorder()
	RequestDiagnostics(mux).ServeHTTP(w, request)
	if w.Code != 503 || w.Header().Get("X-Trace-ID") != traceID {
		t.Fatal("request trace or status changed")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	logs := output.String()
	for _, value := range []string{traceID, "GET /items/{id}", "HTTP dependency", "http.server.request.count", "http.server.request.duration"} {
		if !strings.Contains(logs, value) {
			t.Errorf("missing telemetry: %s", value)
		}
	}
	if strings.Contains(logs, private) {
		t.Fatal("telemetry leaked private request data")
	}
}
