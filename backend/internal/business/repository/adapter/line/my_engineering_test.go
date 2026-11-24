package line

import (
	"context"
	"strings"
	"testing"

	repositoryApplication "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	repositoryApplicationDto "github.com/963sup/line-mini-app-bot/internal/business/repository/application/dto"
	repositoryApplicationPort "github.com/963sup/line-mini-app-bot/internal/business/repository/application/port"
	repositoryDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/repository/domain/entity"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type engineeringCatalog struct {
	repositoryApplicationPort.Store
}

func (engineeringCatalog) Candidates(_ context.Context, user, cursor, _ string) (repositoryApplicationDto.Page, error) {
	if user != "member" {
		return repositoryApplicationDto.Page{}, repositoryDomainEntity.ErrNotFound
	}
	return repositoryApplicationDto.Page{Items: []repositoryDomainEntity.Repository{{ID: strings.Repeat("a", 32), OwnerID: "C" + strings.Repeat("a", 32), Name: "工程 A", Status: "ACTIVE"}, {ID: strings.Repeat("b", 32), OwnerID: "C" + strings.Repeat("b", 32), Name: "工程 B", Status: "ACTIVE"}}}, nil
}

func (engineeringCatalog) IsMember(_ context.Context, group, user string) (bool, error) {
	return group == "C"+strings.Repeat("a", 32) && user == "member", nil
}

func TestMyEngineeringUsesPrivateIdentityAndCurrentMembership(t *testing.T) {
	svc := &repositoryApplication.Service{Store: engineeringCatalog{}, Membership: engineeringCatalog{}}
	msg := webhook.MessageEvent{Mode: "active", ReplyToken: "token", Source: webhook.UserSource{UserId: "member"}, Message: webhook.TextMessageContent{Text: "我的工程"}}
	reply, handled, err := MyEngineering(context.Background(), msg, "https://example.org", svc)
	if err != nil || !handled {
		t.Fatal(err, handled)
	}
	message := reply.Message.(*messaging_api.TemplateMessage)
	columns := message.Template.(*messaging_api.CarouselTemplate).Columns
	if len(columns) != 1 || columns[0].Text != "工程 A" {
		t.Fatal(columns)
	}
	msg.Source = webhook.GroupSource{GroupId: "A", UserId: "member"}
	if _, handled, err = MyEngineering(context.Background(), msg, "https://example.org", nil); handled || err != nil {
		t.Fatal("group request handled", handled, err)
	}
}
