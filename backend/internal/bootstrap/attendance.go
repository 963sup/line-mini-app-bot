package bootstrap

import (
	"fmt"
	"time"
	_ "time/tzdata"

	"cloud.google.com/go/firestore"
	storage "github.com/963sup/line-mini-app-bot/internal/business/attendance/adapter/firestore"
	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
)

type attendanceDependencies struct {
	store attendanceport.Store
	zone  *time.Location
}

func newAttendance(firestoreConfigured bool, getFirestore func() (*firestore.Client, error), channel string) (*attendanceDependencies, error) {
	if !firestoreConfigured || channel == "" {
		return nil, nil
	}
	if !validChannelID(channel) {
		return nil, fmt.Errorf("attendance requires LINE_MESSAGING_API_CHANNEL_ID")
	}
	client, err := getFirestore()
	if err != nil {
		return nil, fmt.Errorf("cannot initialize attendance database")
	}
	zone, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return nil, err
	}
	return &attendanceDependencies{store: &storage.Attendance{Client: client, ChannelID: channel}, zone: zone}, nil
}
