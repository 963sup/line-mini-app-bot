package httptransport

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const telemetryScope = "line-mini-app-bot/http"

// NewTelemetry exports allowlisted application spans and metrics to the existing
// structured log sink. No request bodies, headers, URLs or SDK attributes leave
// this boundary. The caller must shut down providers after draining requests.
func NewTelemetry(logger *slog.Logger) func(context.Context) error {
	logger = logger.With("service", "line-mini-app-bot")
	traces := sdktrace.NewTracerProvider(sdktrace.WithBatcher(spanLogExporter{logger}))
	metrics := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricLogExporter{logger})))
	otel.SetTracerProvider(traces)
	otel.SetMeterProvider(metrics)
	return func(ctx context.Context) error { return errors.Join(traces.Shutdown(ctx), metrics.Shutdown(ctx)) }
}

type spanLogExporter struct{ logger *slog.Logger }

func (e spanLogExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		if span.InstrumentationScope().Name != telemetryScope {
			continue
		}
		e.logger.InfoContext(ctx, "application span", "trace_id", span.SpanContext().TraceID().String(),
			"span_id", span.SpanContext().SpanID().String(), "parent_span_id", span.Parent().SpanID().String(),
			"operation", span.Name(), "duration_ms", span.EndTime().Sub(span.StartTime()).Milliseconds(),
			"status", span.Status().Code.String())
	}
	return ctx.Err()
}

func (spanLogExporter) Shutdown(ctx context.Context) error { return ctx.Err() }

type metricLogExporter struct{ logger *slog.Logger }

func (metricLogExporter) Temporality(sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (metricLogExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.DefaultAggregationSelector(kind)
}

func (e metricLogExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	for _, scope := range data.ScopeMetrics {
		if scope.Scope.Name != telemetryScope {
			continue
		}
		for _, measurement := range scope.Metrics {
			switch value := measurement.Data.(type) {
			case metricdata.Sum[int64]:
				for _, point := range value.DataPoints {
					e.logger.InfoContext(ctx, "application metric", "metric", measurement.Name, "value", point.Value, "labels", safeLabels(point.Attributes))
				}
			case metricdata.Histogram[float64]:
				for _, point := range value.DataPoints {
					e.logger.InfoContext(ctx, "application metric", "metric", measurement.Name, "count", point.Count, "sum", point.Sum, "bounds", point.Bounds, "buckets", point.BucketCounts, "labels", safeLabels(point.Attributes))
				}
			}
		}
	}
	return ctx.Err()
}

func safeLabels(set attribute.Set) map[string]any {
	labels := map[string]any{}
	for _, label := range set.ToSlice() {
		switch string(label.Key) {
		case "http.request.method", "http.route", "http.response.status_code":
			labels[string(label.Key)] = label.Value.AsInterface()
		}
	}
	return labels
}

func (metricLogExporter) ForceFlush(ctx context.Context) error { return ctx.Err() }
func (metricLogExporter) Shutdown(ctx context.Context) error   { return ctx.Err() }

// TraceTransport propagates W3C context only to requests explicitly using it.
// The input request and shared DefaultTransport are never mutated.
type TraceTransport struct{ Base http.RoundTripper }

func (t TraceTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx, span := otel.Tracer(telemetryScope).Start(request.Context(), "HTTP dependency", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	copy := request.Clone(ctx)
	propagation.TraceContext{}.Inject(ctx, propagation.HeaderCarrier(copy.Header))
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	response, err := base.RoundTrip(copy)
	if err != nil || response != nil && response.StatusCode >= 500 {
		span.SetStatus(codes.Error, "dependency failure")
	}
	return response, err
}
