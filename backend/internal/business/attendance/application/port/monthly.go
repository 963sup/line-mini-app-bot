package port

import (
	"context"
	"time"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

type MonthlyReader interface {
	RecordsOverlappingMonth(context.Context, string, string, time.Time) ([]attendanceentity.Record, error)
}
