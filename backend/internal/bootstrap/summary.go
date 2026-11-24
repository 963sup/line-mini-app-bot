package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	httptransport "github.com/963sup/line-mini-app-bot/internal/transport/http"

	"cloud.google.com/go/firestore"
	storage "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/firestore"
	httpadapter "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/http/summarization"
	lineadapter "github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/line"
	"github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/messaging/cloudtasks/summarization"
	"github.com/963sup/line-mini-app-bot/internal/business/summary/adapter/summarizer/summarization"
	summaryApplicationSummarization "github.com/963sup/line-mini-app-bot/internal/business/summary/application/summarization"
	identitygoogle "github.com/963sup/line-mini-app-bot/internal/platform/google"
	lineplatform "github.com/963sup/line-mini-app-bot/internal/platform/line"
	"golang.org/x/oauth2/google"
)

func registerSummary(ctx context.Context, mux *http.ServeMux, bot *lineplatform.Messaging, channel string, firestoreConfigured bool, getFirestore func() (*firestore.Client, error)) (*lineadapter.Summary, func(context.Context) error, error) {
	provider := os.Getenv("SUMMARY_PROVIDER")
	keyName := "GEMINI_API_KEY"
	if provider == "openai" {
		keyName = "OPENAI_API_KEY"
	}
	key, model, queue, target, email := os.Getenv(keyName), os.Getenv("SUMMARY_MODEL"), os.Getenv("SUMMARY_TASK_QUEUE"), os.Getenv("SUMMARY_TASK_URL"), os.Getenv("SUMMARY_TASK_SERVICE_ACCOUNT")
	if provider == "" && key == "" && model == "" && queue == "" && target == "" && email == "" {
		return nil, nil, nil
	}
	if key == "" || !regexp.MustCompile(`^[A-Za-z0-9.-]+$`).MatchString(model) || !regexp.MustCompile(`^projects/[a-z0-9-]+/locations/[a-z0-9-]+/queues/[A-Za-z0-9_-]+$`).MatchString(queue) || email == "" || !firestoreConfigured || !validChannelID(channel) {
		return nil, nil, fmt.Errorf("incomplete or invalid summary configuration")
	}
	u, e := url.Parse(target)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "/internal/summary" {
		return nil, nil, fmt.Errorf("SUMMARY_TASK_URL must be the HTTPS /internal/summary URL")
	}
	modelClient := &http.Client{Transport: httptransport.TraceTransport{}, Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	modelAdapter, err := summarizer.New(provider, key, model, modelClient)
	if err != nil {
		return nil, nil, err
	}
	db, e := getFirestore()
	if e != nil {
		return nil, nil, e
	}
	auth, e := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if e != nil {
		return nil, nil, e
	}
	auth.Timeout = 4 * time.Second
	auth.CheckRedirect = modelClient.CheckRedirect
	if bot == nil {
		return nil, nil, fmt.Errorf("summary requires a LINE messaging token")
	}
	store := &storage.Summary{Client: db, ChannelID: channel}
	service := &summaryApplicationSummarization.Service{Store: store, Model: modelAdapter, Sender: lineadapter.SummaryDelivery{Bot: bot}}
	mux.HandleFunc("POST /internal/summary", httpadapter.SummaryWorker(service, &identitygoogle.Verifier{Audience: target, Email: email}))
	return &lineadapter.Summary{Store: store, Queue: &cloudtasks.Queue{Client: auth, Name: queue, URL: target, ServiceAccount: email}}, func(ctx context.Context) error { return store.RecoverPending(ctx, service.Run) }, nil
}
