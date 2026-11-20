package query

import (
	"context"
	"time"

	attendancedto "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/dto"

	attendanceport "github.com/963sup/line-mini-app-bot/internal/business/attendance/application/port"
	attendanceservice "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/service"
)

func ReadMonthlyReport(ctx context.Context, reader attendanceport.MonthlyReader, user, group string, month time.Time) (attendancedto.MonthlyReport, error) {
	shifts, err := reader.RecordsOverlappingMonth(ctx, user, group, month)
	if err != nil {
		return attendancedto.MonthlyReport{}, err
	}
	return attendancedto.MonthlyReport(attendanceservice.CalculateMonthlyAttendance(shifts, month)), nil
}
