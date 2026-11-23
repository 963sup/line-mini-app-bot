package cloudtasks

import (
	summaryDomainEntitySummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/summarization"
)

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/propagation"

	"github.com/963sup/line-mini-app-bot/internal/transport/http/api/generatedinternal"
)

type Queue struct {
	Client                    *http.Client
	Name, URL, ServiceAccount string
}

func (q *Queue) Enqueue(ctx context.Context, j summaryDomainEntitySummarization.Job) error {
	payload, _ := json.Marshal(generatedinternal.SummaryJob{Group: j.GroupID, Id: j.ID})
	headers := map[string]string{"Content-Type": "application/json"}
	propagation.TraceContext{}.Inject(ctx, propagation.MapCarrier(headers))
	body, _ := json.Marshal(map[string]any{"task": map[string]any{
		"name": q.Name + "/tasks/" + j.ID, "dispatchDeadline": "120s",
		"httpRequest": map[string]any{"url": q.URL, "httpMethod": "POST", "headers": headers, "body": base64.StdEncoding.EncodeToString(payload), "oidcToken": map[string]string{"serviceAccountEmail": q.ServiceAccount, "audience": q.URL}},
	}})
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudtasks.googleapis.com/v2/"+q.Name+"/tasks", bytes.NewReader(body))
	if e != nil {
		return errors.New("cannot create summary task")
	}
	req.Header.Set("Content-Type", "application/json")
	res, e := q.Client.Do(req)
	if e != nil {
		return errors.New("summary queue unavailable")
	}
	// The response is read-only; closing cannot commit application state.
	defer func() { _ = res.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusConflict {
		return errors.New("summary queue rejected task")
	}
	return nil
}
