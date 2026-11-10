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

type leaveDocument struct {
	GroupID   string    `firestore:"groupID"`
	LaunchKey string    `firestore:"launchKey"`
	Kind      string    `firestore:"kind"`
	Start     time.Time `firestore:"start"`
	End       time.Time `firestore:"end"`
	Reason    string    `firestore:"reason"`
	CreatedAt time.Time `firestore:"createdAt"`
}

func (d leaveDocument) domain(id string) attendance.Leave {
	return attendance.Leave{ID: id, GroupID: d.GroupID, Kind: d.Kind, Start: d.Start, End: d.End, Reason: d.Reason, CreatedAt: d.CreatedAt}
}

func (s *Attendance) leaveLaunches() *db.CollectionRef {
	return s.Client.Collection("channels").Doc(s.ChannelID).Collection("leaveLaunches")
}

func (s *Attendance) createLeaveLaunchOperation(ctx context.Context, group string) (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw[:])
	_, err := s.leaveLaunches().Doc(documentKey(token)).Create(ctx, launch{GroupID: group, ExpiresAt: time.Now().UTC().Add(30 * time.Minute)})
	return token, err
}

func (s *Attendance) leaveGroupOperation(ctx context.Context, token string) (string, error) {
	snap, err := s.leaveLaunches().Doc(documentKey(token)).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return "", attendance.ErrExpiredLaunch
	}
	if err != nil {
		return "", err
	}
	var l launch
	if err = snap.DataTo(&l); err != nil {
		return "", err
	}
	if !time.Now().Before(l.ExpiresAt) {
		return "", attendance.ErrExpiredLaunch
	}
	return l.GroupID, nil
}

func (s *Attendance) submitLeaveOperation(ctx context.Context, token, user, key string, proposal attendance.Leave) (attendance.Leave, error) {
	if !attendance.ValidLeaveKey(key) || user == "" || proposal.GroupID == "" {
		return attendance.Leave{}, attendance.ErrInvalidLeave
	}
	proposal, err := attendance.NormalizeLeave(proposal)
	if err != nil {
		return attendance.Leave{}, err
	}
	ref := s.user(user).Collection("leaveRequests").Doc(documentKey(key))
	var result attendance.Leave
	err = s.Client.RunTransaction(ctx, func(_ context.Context, tx *db.Transaction) error {
		snap, err := tx.Get(ref)
		if err == nil {
			var old leaveDocument
			if err = snap.DataTo(&old); err != nil {
				return err
			}
			if old.LaunchKey != documentKey(token) || old.GroupID != proposal.GroupID || old.Kind != proposal.Kind || !old.Start.Equal(proposal.Start) || !old.End.Equal(proposal.End) || old.Reason != proposal.Reason {
				return attendance.ErrLeaveConflict
			}
			result = old.domain(ref.ID)
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
		snap, err = tx.Get(s.leaveLaunches().Doc(documentKey(token)))
		if status.Code(err) == codes.NotFound {
			return attendance.ErrExpiredLaunch
		}
		if err != nil {
			return err
		}
		var l launch
		if err = snap.DataTo(&l); err != nil {
			return err
		}
		if !time.Now().Before(l.ExpiresAt) || l.GroupID != proposal.GroupID {
			return attendance.ErrExpiredLaunch
		}
		d := leaveDocument{GroupID: l.GroupID, LaunchKey: documentKey(token), Kind: proposal.Kind, Start: proposal.Start, End: proposal.End, Reason: proposal.Reason, CreatedAt: time.Now().UTC()}
		result = d.domain(ref.ID)
		return tx.Create(ref, d)
	})
	return result, err
}

func (s *Attendance) recentLeavesOperation(ctx context.Context, user string) ([]attendance.Leave, error) {
	snaps, err := s.user(user).Collection("leaveRequests").OrderBy("createdAt", db.Desc).Limit(20).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	result := make([]attendance.Leave, 0, len(snaps))
	for _, snap := range snaps {
		var d leaveDocument
		if err := snap.DataTo(&d); err != nil {
			return nil, err
		}
		result = append(result, d.domain(snap.Ref.ID))
	}
	return result, nil
}

var _ attendance.LeaveStore = (*Attendance)(nil)

func (s *Attendance) CreateLeaveLaunch(ctx context.Context, group string) (string, error) {
	value, err := s.createLeaveLaunchOperation(ctx, group)
	return value, translateError(err, nil)
}

func (s *Attendance) LeaveGroup(ctx context.Context, token string) (string, error) {
	value, err := s.leaveGroupOperation(ctx, token)
	return value, translateError(err, nil)
}

func (s *Attendance) SubmitLeave(ctx context.Context, token, user, key string, proposal attendance.Leave) (attendance.Leave, error) {
	value, err := s.submitLeaveOperation(ctx, token, user, key, proposal)
	return value, translateError(err, nil)
}

func (s *Attendance) RecentLeaves(ctx context.Context, user string) ([]attendance.Leave, error) {
	value, err := s.recentLeavesOperation(ctx, user)
	return value, translateError(err, nil)
}
