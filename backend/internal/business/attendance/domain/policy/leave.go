package policy

import (
	"strings"
	"time"
	"unicode/utf8"

	attendanceentity "github.com/963sup/line-mini-app-bot/internal/business/attendance/domain/entity"
)

func NormalizeLeave(l attendanceentity.Leave) (attendanceentity.Leave, error) {
	l.Reason = strings.TrimSpace(l.Reason)
	l.Start, l.End = l.Start.UTC(), l.End.UTC()
	if l.Kind != "personal" && l.Kind != "sick" && l.Kind != "annual" && l.Kind != "other" {
		return attendanceentity.Leave{}, attendanceentity.ErrInvalidLeave
	}
	if l.Start.IsZero() || l.Start.Year() < 2000 || l.Start.Year() > 2100 || !l.End.After(l.Start) || l.End.Sub(l.Start) > 366*24*time.Hour || l.Reason == "" || utf8.RuneCountInString(l.Reason) > 500 {
		return attendanceentity.Leave{}, attendanceentity.ErrInvalidLeave
	}
	return l, nil
}
