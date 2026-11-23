package cloudtasks

import (
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

type taskTransport func(*http.Request) (*http.Response, error)

func (f taskTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSummaryQueueUsesStableIDAndOIDC(t *testing.T) {
	q := Queue{Name: "projects/p/locations/r/queues/q", URL: "https://example.test/internal/summary", ServiceAccount: "tasks@example.test"}
	q.Client = &http.Client{Transport: taskTransport(func(r *http.Request) (*http.Response, error) {
		var data struct {
			Task struct {
				Name        string
				HTTPRequest struct {
					URL, Body string
					Headers   map[string]string
					OIDCToken struct{ ServiceAccountEmail, Audience string }
				}
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Fatal(err)
		}
		if data.Task.Name != q.Name+"/tasks/job" || data.Task.HTTPRequest.OIDCToken.Audience != q.URL || data.Task.HTTPRequest.OIDCToken.ServiceAccountEmail != q.ServiceAccount {
			t.Fatal("task identity or authentication changed")
		}
		if data.Task.HTTPRequest.Headers["traceparent"] == "" {
			t.Fatal("task lost producer trace context")
		}
		return &http.Response{StatusCode: 409, Body: io.NopCloser(strings.NewReader("already exists")), Header: make(http.Header)}, nil
	})}
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled}))
	if err := q.Enqueue(ctx, summaryDomainEntitySummarization.Job{ID: "job", GroupID: "g"}); err != nil {
		t.Fatal("duplicate task must be accepted")
	}
}
