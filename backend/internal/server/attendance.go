package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"
	_ "time/tzdata"

	"cloud.google.com/go/firestore"
	"github.com/963sup/line-mini-app-bot/internal/attendance"
	storage "github.com/963sup/line-mini-app-bot/internal/attendance/firestore"
)

type attendanceDependencies struct {
	store attendance.Store
	zone  *time.Location
}

func newAttendance(ctx context.Context, channel string) (*attendanceDependencies, func() error, error) {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if project == "" || channel == "" {
		return nil, func() error { return nil }, nil
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(channel) {
		return nil, nil, fmt.Errorf("attendance requires LINE_MESSAGING_API_CHANNEL_ID")
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	client, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot initialize attendance database")
	}
	zone, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return nil, nil, errors.Join(err, client.Close())
	}
	return &attendanceDependencies{store: &storage.Attendance{Client: client, ChannelID: channel}, zone: zone}, client.Close, nil
}
