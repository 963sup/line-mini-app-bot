package dto

import (
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
	"github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/service"
)

type MonthlyReport struct {
	Start         time.Time
	Records       []entity.Record
	Calculation   service.Result
	CorrectionIDs []string
}
