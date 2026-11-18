package attendance

import (
	"context"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/business/attendance/punch"
	recordmodel "github.com/963sup/line-mini-app-bot/internal/business/attendance/records"
)

type Store interface {
	Punch(ctx context.Context, groupID, userID, eventID, action string, now time.Time, eventTimeMillis int64) (punch.Result, error)
	Recent(ctx context.Context, userID string) ([]recordmodel.Record, error)
}
