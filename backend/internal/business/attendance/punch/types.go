package punch

import recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"

type Result struct {
	Code   string
	Record recordmodel.Record
}
