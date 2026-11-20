package milestonehttp

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	milestoneentity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity/milestones"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStorageFailureResponse(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, status.Error(codes.PermissionDenied, "private storage details"))
	var response struct{ Error struct{ Code string } }
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if w.Code != 503 || response.Error.Code != "unavailable" || strings.Contains(w.Body.String(), "private storage details") {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestMilestoneCompletionPercent(t *testing.T) {
	if view(milestoneentity.Milestone{}, "actor").CompletionPercent != nil {
		t.Fatal("empty milestone has no progress percentage")
	}
	for _, tc := range []struct {
		done, total int64
		want        int
	}{{0, 4, 0}, {1, 8, 13}, {2, 3, 67}, {3, 3, 100}} {
		value := view(milestoneentity.Milestone{CompletedIssueCount: tc.done, IssueCount: tc.total}, "actor").CompletionPercent
		if value == nil || *value != tc.want {
			t.Errorf("%d/%d: %v", tc.done, tc.total, value)
		}
	}
}
