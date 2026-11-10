package httpadapter

import (
	"context"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/attendance"
	"github.com/963sup/line-mini-app-bot/internal/lineui"
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
	return lineui.GroupName(group, "")
}
