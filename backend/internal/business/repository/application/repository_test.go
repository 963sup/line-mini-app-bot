package application

import (
	"context"
	"strings"
	"testing"

	"github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
	"github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
)

type candidatePage struct {
	port.Store
	dto.Page
}

func (p candidatePage) Candidates(context.Context, string, string, string) (dto.Page, error) {
	return p.Page, nil
}

func TestListForOwnerPreservesAuthorizationAndContinuation(t *testing.T) {
	user := "U" + strings.Repeat("a", 32)
	group := "C" + strings.Repeat("b", 32)
	service := Service{Store: candidatePage{Page: dto.Page{Items: []entity.Repository{
		{ID: "group", OwnerID: group, Status: "ACTIVE"}, {ID: "inactive", OwnerID: user, Status: "INACTIVE"}, {ID: "personal", OwnerID: user, Status: "ACTIVE"},
	}, NextCursor: "more"}}}
	page, err := service.List(t.Context(), user, "", user)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "personal" || page.NextCursor != "more" {
		t.Fatalf("unexpected page: %+v, %v", page, err)
	}
	page, err = service.List(t.Context(), "U"+strings.Repeat("c", 32), "", user)
	if err != nil || len(page.Items) != 0 || page.NextCursor != "more" {
		t.Fatalf("unexpected page for another actor: %+v, %v", page, err)
	}
}
