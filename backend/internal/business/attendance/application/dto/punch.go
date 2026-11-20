package dto

import (
	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

type PunchResult struct {
	Code   string
	Record attendanceentity.Record
}
