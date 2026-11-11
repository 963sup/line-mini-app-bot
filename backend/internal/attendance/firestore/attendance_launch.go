package firestore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	db "cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrExpiredLaunch = attendance.ErrExpiredLaunch

type launch struct {
	GroupID   string    `firestore:"groupID"`
	Action    string    `firestore:"action"`
	ExpiresAt time.Time `firestore:"expiresAt"`
}

func (s *Attendance) launches() *db.CollectionRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("attendanceLaunches")
}

func (s *Attendance) CreateLaunch(ctx context.Context, group, action string) (_ string, err error) {
	defer func() { err = translateError(err, nil) }()
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw[:])
	_, err = s.launches().Doc(documentKey(token)).Create(ctx, launch{GroupID: group, Action: action, ExpiresAt: time.Now().UTC().Add(10 * time.Minute)})
	return token, err
}

// ExchangeLaunch creates a stable request for this card, action and user.
func (s *Attendance) ExchangeLaunch(ctx context.Context, token, user string) (_ string, err error) {
	defer func() { err = translateError(err, nil) }()
	id := documentKey(token, user)
	err = s.Client.RunTransaction(ctx, func(ctx context.Context, tx *db.Transaction) error {
		snap, err := tx.Get(s.launches().Doc(documentKey(token)))
		if status.Code(err) == codes.NotFound {
			return ErrExpiredLaunch
		}
		if err != nil {
			return err
		}
		var l launch
		if err = snap.DataTo(&l); err != nil {
			return err
		}
		if !time.Now().UTC().Before(l.ExpiresAt) {
			return ErrExpiredLaunch
		}
		ref := s.requests().Doc(id)
		_, err = tx.Get(ref)
		if err == nil {
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		return tx.Create(ref, requestDocument{GroupID: l.GroupID, UserID: documentKey(user), EventID: documentKey(token, user), Action: l.Action, EventTime: time.Now().UnixMilli(), ExpiresAt: l.ExpiresAt})
	})
	return id, err
}

func (s *Attendance) LaunchGroup(ctx context.Context, token string) (_ string, err error) {
	defer func() { err = translateError(err, nil) }()
	snap, err := s.launches().Doc(documentKey(token)).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return "", ErrExpiredLaunch
	}
	if err != nil {
		return "", err
	}
	var l launch
	if err = snap.DataTo(&l); err != nil {
		return "", err
	}
	if !time.Now().UTC().Before(l.ExpiresAt) {
		return "", ErrExpiredLaunch
	}
	return l.GroupID, nil
}
