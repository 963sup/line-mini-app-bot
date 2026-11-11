package attendancehttp

import "net/http"

func RegisterAttendanceAPI(public *http.ServeMux, bot *Attendance, app Authenticator) {
	registerAttendanceAPIWithIdentity(public, bot, app)
}

func registerAttendanceAPIWithIdentity(public *http.ServeMux, bot *Attendance, app Authenticator) {
	mux := http.NewServeMux()
	registerAttendanceQueries(mux, bot, app)
	registerAttendanceCorrections(mux, bot, app)
	registerAttendanceLaunch(mux, bot, app)
	registerPunchPreflight(mux, bot, app)
	registerLeave(mux, bot, app)
	registerPunch(mux, bot, app)
	registerContractAttendance(public, mux, bot, app)
}
