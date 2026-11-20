package projecthttp

import (
	projectDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/project/domain/entity"
)

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/identity"

	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

func inspectionView(v projectDomainEntity.InspectionSubmission) generated.InspectionSubmission {
	return generated.InspectionSubmission{Id: v.ID, Note: v.Note, SubmittedAt: v.SubmittedAt, SubmittedBy: v.SubmittedBy, SourceId: v.SourceID, SourceType: v.SourceType, SourceRepositoryId: v.SourceRepositoryID, SourceVersion: v.SourceVersion, Status: generated.InspectionSubmissionStatus(v.Status)}
}

func (s *projectServer) ListInspectionSubmissions(w http.ResponseWriter, r *http.Request, sid, pid, item string) {
	entries, err := s.service.Inspections(r.Context(), identity.UserID(r.Context()), sid, pid, item, nil)
	if !writeError(w, err) {
		views := make([]generated.InspectionSubmission, 0, len(entries))
		for _, entry := range entries {
			views = append(views, inspectionView(entry))
		}
		transport.WriteJSON(w, http.StatusOK, views)
	}
}

func (s *projectServer) SubmitInspection(w http.ResponseWriter, r *http.Request, sid, pid, item string, _ generated.SubmitInspectionParams) {
	var input generated.InspectionInput
	if !transport.ReadJSON(w, r, &input, "資料無效。") {
		return
	}
	entries, err := s.service.Inspections(r.Context(), identity.UserID(r.Context()), sid, pid, item, &input.Note)
	if !writeError(w, err) {
		transport.WriteJSON(w, http.StatusOK, inspectionView(entries[0]))
	}
}
