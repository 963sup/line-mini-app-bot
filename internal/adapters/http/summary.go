package httpadapter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/963sup/line-mini-app-bot/internal/summary"
	"google.golang.org/api/idtoken"
)

func SummaryWorker(service *summary.Service, target, email string) http.HandlerFunc {
	groupIDPattern := regexp.MustCompile(`^C[0-9a-f]{32}$`)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 80*time.Second)
		defer cancel()
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		claims, e := idtoken.Validate(ctx, parts[1], target)
		if e != nil || claims.Claims["email"] != email || claims.Claims["email_verified"] != true {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var input struct{ Group, ID string }
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048))
		decoder.DisallowUnknownFields()
		if e := decoder.Decode(&input); e != nil || !groupIDPattern.MatchString(input.Group) || !requestIDPattern.MatchString(input.ID) {
			http.Error(w, "invalid job", http.StatusBadRequest)
			return
		}
		if decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "invalid job", http.StatusBadRequest)
			return
		}
		if e := service.Run(ctx, input.Group, input.ID); e != nil {
			http.Error(w, "summary temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
