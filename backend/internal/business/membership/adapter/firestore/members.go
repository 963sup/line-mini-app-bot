package membershipdb

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/business/membership/application"
	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (w work) Roster(id string) (application.Roster, error) {
	record, err := w.readOrganization(id)
	return application.Roster{Name: record.Name, OwnerID: record.Owner, UserIDs: record.Members}, err
}

func (w work) departureReceipt(org, actor, request string) *firestore.DocumentRef {
	return w.root.Collection("membershipAudits").Doc(firestorekey.Encode("departure", org, actor, request))
}

func (w work) Departure(org, actor, request string) (application.Departure, bool, error) {
	snap, err := w.tx.Get(w.departureReceipt(org, actor, request))
	if status.Code(err) == codes.NotFound {
		return application.Departure{}, false, nil
	}
	if err != nil {
		return application.Departure{}, false, err
	}
	var record struct {
		TargetRef string `firestore:"targetRef"`
	}
	err = snap.DataTo(&record)
	return application.Departure{TargetRef: record.TargetRef}, true, err
}

func (w work) departureMarker(org, actor string) *firestore.DocumentRef {
	return w.organization(org).Collection("membershipDepartures").Doc(firestorekey.Encode(actor))
}

func (w work) LastDeparture(org, actor string) (time.Time, error) {
	snap, err := w.tx.Get(w.departureMarker(org, actor))
	if status.Code(err) == codes.NotFound {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	var record struct {
		At time.Time `firestore:"at"`
	}
	err = snap.DataTo(&record)
	return record.At, err
}

func (w work) SaveDeparture(org, target string, d application.Departure) error {
	if err := w.tx.Update(w.organization(org), []firestore.Update{{Path: "memberUserIds", Value: firestore.ArrayRemove(target)}}); err != nil {
		return err
	}
	if err := w.tx.Set(w.departureMarker(org, target), map[string]any{"at": d.At}); err != nil {
		return err
	}
	return w.tx.Create(w.departureReceipt(org, d.ActorID, d.RequestID), map[string]any{
		"resourceType": "organization", "resourceId": org, "actorUserId": d.ActorID,
		"targetUserId": target, "targetRef": d.TargetRef, "requestId": d.RequestID,
		"at": d.At, "action": "end_membership",
	})
}
