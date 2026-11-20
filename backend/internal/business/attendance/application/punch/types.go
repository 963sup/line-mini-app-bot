package punch

import recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/records"

type Result struct {
	Code   string
	Record recordmodel.Record
}
