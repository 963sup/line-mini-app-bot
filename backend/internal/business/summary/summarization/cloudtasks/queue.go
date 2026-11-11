package cloudtasks

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/business/summary/summarization"
)

type Queue struct {
	Client                    *http.Client
	Name, URL, ServiceAccount string
}

func (q *Queue) Enqueue(ctx context.Context, j summarization.Job) error {
	payload, _ := json.Marshal(map[string]string{"group": j.GroupID, "id": j.ID})
	body, _ := json.Marshal(map[string]any{"task": map[string]any{
		"name": q.Name + "/tasks/" + j.ID, "dispatchDeadline": "120s",
		"httpRequest": map[string]any{"url": q.URL, "httpMethod": "POST", "headers": map[string]string{"Content-Type": "application/json"}, "body": base64.StdEncoding.EncodeToString(payload), "oidcToken": map[string]string{"serviceAccountEmail": q.ServiceAccount, "audience": q.URL}},
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
