package httptransport

import (
	"crypto/rand"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// RequestDiagnostics assigns a server-generated correlation ID. Never trust an
// incoming ID or include URL paths, queries, headers or bodies in failure logs.
func RequestDiagnostics(next http.Handler) http.Handler {
	meter := otel.Meter(telemetryScope)
	requests, _ := meter.Int64Counter("http.server.request.count")
	duration, _ := meter.Float64Histogram("http.server.request.duration", metric.WithUnit("s"), metric.WithExplicitBucketBoundaries(.01, .05, .1, .5, 1, 5, 15, 30, 90))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := otel.Tracer(telemetryScope).Start(ctx, "HTTP request", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		w.Header().Set("X-Request-ID", rand.Text())
		w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())
		response := &diagnosticResponse{ResponseWriter: w, status: http.StatusOK}
		r = r.WithContext(ctx)
		defer func() {
			failure := recover()
			if failure != nil {
				response.status = http.StatusInternalServerError
			}
			route := r.Pattern
			if _, pattern, found := strings.Cut(route, " "); found {
				route = pattern // ServeMux patterns may include the HTTP method.
			}
			if route == "" {
				route = "unmatched"
			}
			method := r.Method
			switch method {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
			default:
				method = "OTHER"
			}
			labels := []attribute.KeyValue{attribute.String("http.request.method", method), attribute.String("http.route", route), attribute.Int("http.response.status_code", response.status)}
			span.SetName(method + " " + route)
			span.SetAttributes(labels...)
			if response.status >= 500 {
				span.SetStatus(codes.Error, "server failure")
			}
			requests.Add(ctx, 1, metric.WithAttributes(labels...))
			duration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(labels...))
			if failure != nil {
				panic(failure)
			}
		}()
		next.ServeHTTP(response, r)
	})
}

type diagnosticResponse struct {
	http.ResponseWriter
	status  int
	written bool
}

func (w *diagnosticResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *diagnosticResponse) WriteHeader(status int) {
	if status >= 200 && !w.written {
		w.status = status
		w.written = true
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *diagnosticResponse) Write(body []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
func (w *diagnosticResponse) Flush() { _ = w.FlushError() }
func (w *diagnosticResponse) FlushError() error {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}
