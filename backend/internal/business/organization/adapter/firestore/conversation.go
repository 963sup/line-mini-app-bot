package organizationdb

import (
	"context"
	"errors"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	enterprisedb "github.com/963sup/line-mini-app-bot/internal/business/enterprise/adapter/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/internal/adapter/firestore"
	dto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
	organizationDomainEntity "github.com/963sup/line-mini-app-bot/internal/business/organization/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	pendingCreation struct {
		Kind    string
		Started int64
		Expires time.Time
	}
)

// Conversation commits the resource, prompt state and replay receipt together.
func (s Store) Conversation(ctx context.Context, user, eventID, text string, at time.Time) (*dto.CreationReply, error) {
	text = strings.TrimSpace(text)
	now := time.Now()
	id := uuid.NewString()
	return orgstore.Run(ctx, s.DB, s.Channel, func(tx *firestore.Transaction, root *firestore.DocumentRef) (*dto.CreationReply, error) {
		state := root.Collection("Conversations").Doc(firestorekey.Digest([]byte(user)))
		receipt := state.Collection("events").Doc(firestorekey.Digest([]byte(eventID)))
		var pending pendingCreation
		snap, e := tx.Get(state)
		if e == nil {
			if err := snap.DataTo(&pending); err != nil {
				return nil, err
			}
		} else if status.Code(e) != codes.NotFound {
			return nil, e
		} else if text != "建立組織" && text != "建立企業" {
			return nil, nil
		}
		snap, e = tx.Get(receipt)
		if e == nil {
			var reply dto.CreationReply
			e = snap.DataTo(&reply)
			return &reply, e
		}
		if status.Code(e) != codes.NotFound {
			return nil, e
		}
		if at.UnixMilli() < pending.Started {
			return nil, nil
		}
		reply := dto.CreationReply{}
		switch {
		case text == "建立組織" || text == "建立企業":
			kind := "organization"
			label := "組織"
			if text == "建立企業" {
				kind = "enterprise"
				label = "企業"
			}
			pending = pendingCreation{Started: at.UnixMilli()}
			if !at.Add(5 * time.Minute).After(now) {
				reply.Text = "建立流程已逾時，請重新輸入「" + text + "」。"
				break
			}
			if kind == "enterprise" {
				g, _, err := enterprisedb.Provisioning(tx, root, user)
				if err != nil {
					return nil, err
				}
				if !g.Enabled || g.EnterpriseID != "" {
					reply.Text = "帳號目前沒有可用的企業開通資格。"
					break
				}
			}
			pending.Kind = kind
			pending.Expires = at.Add(5 * time.Minute)
			reply.Text = "新" + label + "叫什麼名字？請在五分鐘內回覆名稱，或輸入「取消」。"
		case pending.Kind == "":
			return nil, nil
		case !pending.Expires.After(now):
			reply.Text = "建立流程已逾時，請重新輸入建立指令。"
			pending.Kind = ""
		case text == "取消":
			reply.Text = "已取消建立。"
			pending.Kind = ""
		case text == "" || len([]rune(text)) > 120:
			reply.Text = "請輸入 1 到 120 字的名稱，或輸入「取消」。"
		default:
			kind := pending.Kind
			reply.ID = id
			name := text
			if kind == "enterprise" {
				v, err := enterprisedb.CreateInTransaction(tx, root, user, id, text, false)
				if errors.Is(err, organizationDomainEntity.ErrForbidden) {
					reply = dto.CreationReply{Text: "帳號目前沒有可用的企業開通資格。"}
					pending.Kind = ""
					break
				}
				if err != nil {
					return nil, err
				}
				reply.ID = v.ID
				name = v.Name
			} else {
				if _, err := createOrganization(tx, root, user, id, text); err != nil {
					return nil, err
				}
			}
			reply.Text = "已建立「" + name + "」。"
			reply.Kind = kind
			pending.Kind = ""
		}
		pending.Started = at.UnixMilli()
		if err := tx.Set(state, pending); err != nil {
			return nil, err
		}
		return &reply, tx.Create(receipt, reply)
	})
}
