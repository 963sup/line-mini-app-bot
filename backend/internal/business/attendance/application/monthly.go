package application

import (
	"context"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/records"
)

type MonthlyReader interface {
	RecordsOverlappingMonth(context.Context, string, string, time.Time) ([]records.Record, error)
}

func ReadMonthlyReport(ctx context.Context, reader MonthlyReader, user, group string, month time.Time) (records.MonthlyReport, error) {
	shifts, err := reader.RecordsOverlappingMonth(ctx, user, group, month)
	if err != nil {
		return records.MonthlyReport{}, err
	}
	return records.CalculateMonthlyAttendance(shifts, month), nil
}
