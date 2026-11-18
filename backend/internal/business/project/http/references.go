package projecthttp

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/identity"

	feature "github.com/963sup/line-mini-app-bot/internal/business/project"
	"github.com/963sup/line-mini-app-bot/internal/httpapi/generated"
	transport "github.com/963sup/line-mini-app-bot/internal/transport/http"
)

func referenceView(ref feature.Reference) generated.ProjectReference {
	out := generated.ProjectReference{Id: ref.ID, SourceType: generated.ProjectReferenceSourceType(ref.SourceType), SourceId: ref.SourceID, SourceRepositoryId: ref.SourceRepositoryID, SourceUnavailable: ref.SourceUnavailable, SourceTitle: ref.SourceTitle, SourceState: ref.SourceState}
	if ref.ParentObjectID != "" {
		out.ParentObjectId = &ref.ParentObjectID
	}
	if !ref.SourceUnavailable {
		if ref.Quantity != "" {
			out.Quantity = &ref.Quantity
		}
		if ref.Unit != "" {
			out.Unit = &ref.Unit
		}
		if ref.CompletedQuantity != "" {
			out.CompletedQuantity = &ref.CompletedQuantity
		}
		if ref.Type != "" {
			out.Type = &ref.Type
		}
	}
	return out
}

func (s *projectServer) ListProjectReferences(w http.ResponseWriter, r *http.Request, sid, pid string, p generated.ListProjectReferencesParams) {
	cursor := ""
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, e := s.service.ListReferences(r.Context(), identity.UserID(r.Context()), sid, pid, cursor)
	if writeError(w, e) {
		return
	}
	items := make([]generated.ProjectReference, 0, len(page.Items))
	for _, ref := range page.Items {
		items = append(items, referenceView(ref))
	}
	transport.WriteJSON(w, 200, generated.ProjectReferencePage{Items: items, NextCursor: page.NextCursor})
}

func (s *projectServer) AddProjectReference(w http.ResponseWriter, r *http.Request, sid, pid string, _ generated.AddProjectReferenceParams) {
	var input generated.ProjectReferenceAddition
	if !transport.ReadJSON(w, r, &input, "資料無效。") {
		return
	}
	ref := feature.Reference{SourceType: string(input.SourceType), SourceID: input.SourceId}
	if input.SourceRepositoryId != nil {
		ref.SourceRepositoryID = *input.SourceRepositoryId
	}
	if input.ParentObjectId != nil {
		ref.ParentObjectID = *input.ParentObjectId
	}
	value, e := s.service.AddReference(r.Context(), identity.UserID(r.Context()), sid, pid, ref, input.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(value))
	}
}

func (s *projectServer) RemoveProjectReference(w http.ResponseWriter, r *http.Request, sid, pid, itemID string, p generated.RemoveProjectReferenceParams) {
	value, e := s.service.RemoveReference(r.Context(), identity.UserID(r.Context()), sid, pid, itemID, p.ExpectedVersion)
	if !writeError(w, e) {
		transport.WriteJSON(w, 200, view(value))
	}
}
