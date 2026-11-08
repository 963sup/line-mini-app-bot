package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	storage "github.com/963sup/line-mini-app-bot/internal/adapters/firestore"
	"github.com/963sup/line-mini-app-bot/internal/summary"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type summaryBot struct {
	store summary.Store
	queue summary.Queue
}
type summaryTasks struct {
	client            *http.Client
	queue, url, email string
}
type summaryDelivery struct {
	bot *messaging_api.MessagingApiAPI
}

func (d summaryDelivery) Send(ctx context.Context, group, text, retry string) error {
	client := *d.bot
	client.WithContext(ctx)
	res, _, e := client.PushMessageWithHttpInfo(&messaging_api.PushMessageRequest{To: group, Messages: []messaging_api.MessageInterface{&messaging_api.TextMessage{Text: text}}}, retry)
	// LINE's 409 means this retry key was already accepted, not a new delivery.
	if res != nil && res.StatusCode == http.StatusConflict {
		return nil
	}
	return e
}

func (q *summaryTasks) Enqueue(ctx context.Context, j summary.Job) error {
	payload, _ := json.Marshal(map[string]string{"group": j.GroupID, "id": j.ID})
	body, _ := json.Marshal(map[string]any{"task": map[string]any{
		"name": q.queue + "/tasks/" + j.ID, "dispatchDeadline": "120s",
		"httpRequest": map[string]any{"url": q.url, "httpMethod": "POST", "headers": map[string]string{"Content-Type": "application/json"}, "body": base64.StdEncoding.EncodeToString(payload), "oidcToken": map[string]string{"serviceAccountEmail": q.email, "audience": q.url}},
	}})
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudtasks.googleapis.com/v2/"+q.queue+"/tasks", bytes.NewReader(body))
	if e != nil {
		return errors.New("cannot create summary task")
	}
	req.Header.Set("Content-Type", "application/json")
	res, e := q.client.Do(req)
	if e != nil {
		return errors.New("summary queue unavailable")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusConflict {
		return errors.New("summary queue rejected task")
	}
	return nil
}

func registerSummary(ctx context.Context, mux *http.ServeMux) (*summaryBot, func() error, error) {
	noop := func() error { return nil }
	key, model, queue, target, email := os.Getenv("GEMINI_API_KEY"), os.Getenv("SUMMARY_MODEL"), os.Getenv("SUMMARY_TASK_QUEUE"), os.Getenv("SUMMARY_TASK_URL"), os.Getenv("SUMMARY_TASK_SERVICE_ACCOUNT")
	if key == "" && model == "" && queue == "" && target == "" && email == "" {
		return nil, noop, nil
	}
	project, channel := os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("LINE_MESSAGING_API_CHANNEL_ID")
	if key == "" || !regexp.MustCompile(`^[A-Za-z0-9.-]+$`).MatchString(model) || !regexp.MustCompile(`^projects/[a-z0-9-]+/locations/[a-z0-9-]+/queues/[A-Za-z0-9_-]+$`).MatchString(queue) || email == "" || project == "" || !regexp.MustCompile(`^[0-9]+$`).MatchString(channel) {
		return nil, nil, fmt.Errorf("incomplete or invalid summary configuration")
	}
	u, e := url.Parse(target)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "/internal/summary" {
		return nil, nil, fmt.Errorf("SUMMARY_TASK_URL must be the HTTPS /internal/summary URL")
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	db, e := firestore.NewClientWithDatabase(ctx, project, database)
	if e != nil {
		return nil, nil, errors.New("cannot initialize summary database")
	}
	auth, e := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if e != nil {
		db.Close()
		return nil, nil, errors.New("cannot initialize summary queue credentials")
	}
	auth.Timeout = 4 * time.Second
	noRedirect := func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	auth.CheckRedirect = noRedirect
	line, e := messaging_api.NewMessagingApiAPI(os.Getenv("LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN"), messaging_api.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))
	if e != nil {
		db.Close()
		return nil, nil, e
	}
	store := &storage.Chat{Client: db, ChannelID: channel}
	service := &summary.Service{Store: store, Model: &summary.Gemini{Key: key, Model: model, HTTP: &http.Client{Timeout: 45 * time.Second, CheckRedirect: noRedirect}}, Sender: summaryDelivery{bot: line}}
	mux.HandleFunc("POST /internal/summary", summaryWorker(service, target, email))
	return &summaryBot{store: store, queue: &summaryTasks{client: auth, queue: queue, url: target, email: email}}, db.Close, nil
}

func summaryWorker(service *summary.Service, target, email string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 80*time.Second)
		defer cancel()
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", 401)
			return
		}
		claims, e := idtoken.Validate(ctx, parts[1], target)
		if e != nil || claims.Claims["email"] != email || claims.Claims["email_verified"] != true {
			http.Error(w, "unauthorized", 401)
			return
		}
		var input struct{ Group, ID string }
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048))
		decoder.DisallowUnknownFields()
		if e := decoder.Decode(&input); e != nil || !regexp.MustCompile(`^C[0-9a-f]{32}$`).MatchString(input.Group) || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(input.ID) {
			http.Error(w, "invalid job", 400)
			return
		}
		if decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "invalid job", 400)
			return
		}
		if e := service.Run(ctx, input.Group, input.ID); e != nil {
			http.Error(w, "summary temporarily unavailable", 503)
			return
		}
		w.WriteHeader(204)
	}
}

// observe is independent of command replies: ordinary text remains silent.
func (b *summaryBot) observe(ctx context.Context, event webhook.EventInterface) (lineReply, bool, error) {
	if b == nil {
		return lineReply{}, false, nil
	}
	if e, ok := event.(webhook.MessageEditedEvent); ok {
		g, group := e.Source.(webhook.GroupSource)
		m, text := e.Message.(webhook.TextMessageContent)
		if editor, supported := b.store.(interface {
			Edit(context.Context, string, string, string, int64) error
		}); supported && group && text && g.GroupId != "" && e.Mode == "active" {
			return lineReply{}, false, editor.Edit(ctx, g.GroupId, m.Id, m.Text, e.Timestamp)
		}
		return lineReply{}, false, nil
	}
	if e, ok := event.(webhook.UnsendEvent); ok {
		g, ok := e.Source.(webhook.GroupSource)
		if !ok || g.GroupId == "" || e.Unsend == nil {
			return lineReply{}, false, nil
		}
		return lineReply{}, false, b.store.Unsend(ctx, g.GroupId, e.Unsend.MessageId)
	}
	e, ok := event.(webhook.MessageEvent)
	if !ok || e.Mode != "active" {
		return lineReply{}, false, nil
	}
	g, ok := e.Source.(webhook.GroupSource)
	if !ok || g.GroupId == "" {
		return lineReply{}, false, nil
	}
	m, ok := e.Message.(webhook.TextMessageContent)
	if !ok {
		return lineReply{}, false, nil
	}
	command := strings.TrimSpace(m.Text)
	if command == "摘要" || command == "關閉摘要" || command == "開啟摘要" {
		if time.Since(time.UnixMilli(e.Timestamp)) > 24*time.Hour || time.UnixMilli(e.Timestamp).After(time.Now().Add(time.Minute)) {
			return lineReply{}, true, nil
		}
		if g.UserId == "" || e.WebhookEventId == "" || e.ReplyToken == "" {
			return lineReply{}, true, nil
		}
		j, err := b.store.Command(ctx, g.GroupId, e.WebhookEventId, command, e.Timestamp)
		if err != nil {
			return lineReply{}, true, err
		}
		if j.Status == "done" {
			return lineReply{}, true, nil
		}
		text := "正在整理最近24小時的群聊（最多100則）。"
		if j.Status == "disabled" {
			text = "摘要已關閉，輸入「開啟摘要」可恢復。"
		} else if j.Status != "done" {
			if err := b.queue.Enqueue(ctx, j); err != nil {
				return lineReply{}, true, err
			}
		}
		if command == "關閉摘要" {
			text = "摘要已關閉，既有資料已停止使用並排程清除。"
		}
		if command == "開啟摘要" {
			text = "摘要已開啟。"
		}
		return lineReply{token: e.ReplyToken, message: &messaging_api.TextMessage{Text: text}}, true, nil
	}
	if skipSummaryText(command) {
		return lineReply{}, false, nil
	}
	return lineReply{}, false, b.store.Collect(ctx, g.GroupId, summary.Message{ID: m.Id, EventID: e.WebhookEventId, UserID: g.UserId, Text: m.Text, OccurredAt: time.UnixMilli(e.Timestamp).UTC()})
}

func skipSummaryText(text string) bool {
	if strings.HasPrefix(text, "/") {
		return true
	}
	switch text {
	case "上班", "下班", "打卡", "出勤紀錄", "選單", "功能", "":
		return true
	}
	return false
}
