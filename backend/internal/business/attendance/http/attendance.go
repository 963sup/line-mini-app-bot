package attendancehttp

import (
	"context"
	"net/http"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/presentation"
)

// Attendance contains the dependencies needed by browser attendance handlers.
type Attendance struct {
	store     attendance.Store
	zone      *time.Location
	groupName func(context.Context, string) string
}

func NewAttendance(store attendance.Store, zone *time.Location, groupName func(context.Context, string) string) *Attendance {
	return &Attendance{store: store, zone: zone, groupName: groupName}
}

func (b *Attendance) format(t time.Time) string { return t.In(b.zone).Format("2006-01-02 15:04:05") }

func (b *Attendance) historyGroupName(ctx context.Context, group string) string {
	if b.groupName != nil {
		return b.groupName(ctx, group)
	}
	return presentation.GroupName(group, "")
}

type Authenticator interface {
	Authenticate(http.ResponseWriter, *http.Request) (string, string, bool)
}
