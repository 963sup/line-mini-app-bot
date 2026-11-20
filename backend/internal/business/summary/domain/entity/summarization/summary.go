package summarization

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	summaryDomainEntityMessagecollection "github.com/963sup/line-mini-app-bot/internal/business/summary/domain/entity/messagecollection"
)

type Job struct {
	ID, GroupID, Kind, Result, RetryKey, Status string
	Epoch, Revision                             int64
	Cutoff, ExpireAt, LeaseUntil                time.Time
	Lease                                       string
	// Sources contain identifiers only; no additional copy of chat content.
	Sources []string
}

// RandomID supplies a stable persisted LINE retry key or a unique claim token.
func RandomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random source unavailable")
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	s := hex.EncodeToString(b[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

// Input aliases users without losing the source-message association held by the job.
// The byte budget also bounds the maximum possible input token count.
func Input(messages []summaryDomainEntityMessagecollection.Message) (string, []string) {
	people := map[string]int{}
	lines := make([]string, 0, len(messages))
	sources := make([]string, 0, len(messages))
	budget := 24000
	for _, m := range messages {
		if len(sources) == 100 {
			break
		}
		if m.Deleted || m.Text == "" {
			continue
		}
		if _, ok := people[m.UserID]; !ok {
			people[m.UserID] = len(people) + 1
		}
		b, _ := json.Marshal(struct{ Source, Speaker, Time, Text string }{fmt.Sprintf("S%d", len(sources)+1), fmt.Sprintf("成員%d", people[m.UserID]), m.OccurredAt.UTC().Format(time.RFC3339), m.Text})
		if len(b) > budget {
			break
		}
		budget -= len(b)
		lines = append(lines, string(b))
		sources = append(sources, m.ID)
	}
	return strings.Join(lines, "\n"), sources
}
