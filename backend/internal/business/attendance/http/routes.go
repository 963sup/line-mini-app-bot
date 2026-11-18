package attendancehttp

import (
	"net/http"

	"github.com/963sup/line-mini-app-bot/internal/httpapi"
)

type attendanceServer struct {
	bot *Attendance
	app Authenticator
}

var _ httpapi.AttendanceServer = (*attendanceServer)(nil)

func RegisterAttendanceAPI(public *http.ServeMux, bot *Attendance, app Authenticator) {
	registerAttendanceAPIWithIdentity(public, bot, app)
}

func registerAttendanceAPIWithIdentity(public *http.ServeMux, bot *Attendance, app Authenticator) {
	registerContractAttendance(public, &attendanceServer{bot: bot, app: app}, bot, app)
}
