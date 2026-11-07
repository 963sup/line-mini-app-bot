package summary

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type memoryStore struct {
	job         Job
	messages    []Message
	current     bool
	claimed     bool
	saved, done int
}

func (s *memoryStore) Collect(context.Context, string, Message) error { return nil }
func (s *memoryStore) Unsend(context.Context, string, string) error   { return nil }
func (s *memoryStore) Command(context.Context, string, string, string, int64) (Job, error) {
	return s.job, nil
}
func (s *memoryStore) Claim(context.Context, string, string) (Job, bool, error) {
	return s.job, s.claimed, nil
}
func (s *memoryStore) Messages(context.Context, Job) ([]Message, error) { return s.messages, nil }
func (s *memoryStore) SaveResult(_ context.Context, j Job, text string, ids []string) error {
	s.job = j
	s.job.Result = text
	s.job.Sources = ids
	s.saved++
	return nil
}
func (s *memoryStore) Current(context.Context, Job) (bool, error) { return s.current, nil }
func (s *memoryStore) Done(context.Context, Job) error            { s.done++; return nil }
func (s *memoryStore) Purge(context.Context, Job) error           { return nil }

type modelStub struct {
	calls int
	after func()
}

func (m *modelStub) Summarize(context.Context, []Message) (string, error) {
	m.calls++
	if m.after != nil {
		m.after()
	}
	return "討論重點：待確認 [S1]", nil
}

type senderStub struct {
	calls int
	err   error
	retry string
}

func (s *senderStub) Send(_ context.Context, _, _, retry string) error {
	s.calls++
	s.retry = retry
	return s.err
}

func TestRunReusesSavedResultAfterPushFailure(t *testing.T) {
	now := time.Now()
	store := &memoryStore{claimed: true, current: true, job: Job{ID: "job", RetryKey: "stable", ExpireAt: now.Add(time.Hour)}, messages: []Message{{ID: "message", UserID: "user", Text: "討論", ExpireAt: now.Add(time.Hour)}}}
	model := &modelStub{}
	sender := &senderStub{err: errors.New("timeout")}
	service := Service{Store: store, Model: model, Sender: sender}
	if service.Run(context.Background(), "g", "job") == nil || model.calls != 1 || store.saved != 1 || store.done != 0 {
		t.Fatal("must persist before push and retain failed job")
	}
	sender.err = nil
	if err := service.Run(context.Background(), "g", "job"); err != nil || model.calls != 1 || sender.calls != 2 || sender.retry != "stable" || store.done != 1 {
		t.Fatal("retry must reuse result and retry key")
	}
}
func TestRunCancelsIfDisabledDuringGeneration(t *testing.T) {
	now := time.Now()
	store := &memoryStore{claimed: true, current: true, job: Job{ExpireAt: now.Add(time.Hour)}, messages: []Message{{ID: "m", UserID: "u", Text: "text", ExpireAt: now.Add(time.Hour)}}}
	model := &modelStub{after: func() { store.current = false }}
	sender := &senderStub{}
	service := Service{Store: store, Model: model, Sender: sender}
	if err := service.Run(context.Background(), "g", "id"); err != nil || sender.calls != 0 || store.done != 1 {
		t.Fatal("changed generation must never send")
	}
}
func TestRunDuplicateAndCancelledJobsDoNotCallModel(t *testing.T) {
	for _, claimed := range []bool{false, true} {
		store := &memoryStore{claimed: claimed}
		model := &modelStub{}
		sender := &senderStub{}
		service := Service{Store: store, Model: model, Sender: sender}
		if err := service.Run(context.Background(), "g", "id"); err != nil || model.calls != 0 || sender.calls != 0 {
			t.Fatal("duplicate or cancelled work must stop")
		}
	}
}
func TestInputPreservesActorAndSourcesWithoutRawIDs(t *testing.T) {
	text, ids := Input([]Message{{ID: "m1", UserID: "secret-user", Text: "一"}, {ID: "m2", UserID: "other-user", Text: "二"}, {ID: "m3", UserID: "secret-user", Text: "三"}, {ID: "m4", Text: "撤回", Deleted: true}})
	if len(ids) != 3 || ids[2] != "m3" || strings.Contains(text, "secret-user") || strings.Count(text, "成員1") != 2 || !strings.Contains(text, "S3") || strings.Contains(text, "撤回") {
		t.Fatal("actor aliases or citations are inconsistent")
	}
	_, ids = Input([]Message{{ID: "huge", Text: strings.Repeat("x", 25000)}})
	if len(ids) != 0 {
		t.Fatal("input budget must be bounded")
	}
}
