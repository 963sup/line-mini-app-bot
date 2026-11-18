package httptransport

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSON(t *testing.T) {
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{"{\"id\":1}  ", true}, {"", false}, {"{\"other\":1}", false}, {"{} {}", false}, {"{} !", false},
	} {
		t.Run(tc.body, func(t *testing.T) {
			var target struct{ ID int }
			w := httptest.NewRecorder()
			ok := ReadJSON(w, httptest.NewRequest("POST", "/", strings.NewReader(tc.body)), &target, "invalid draft")
			if ok != tc.valid {
				t.Fatalf("accepted = %v", ok)
			}
			if ok {
				if target.ID != 1 {
					t.Fatal(target)
				}
				return
			}
			var response struct {
				Error struct{ Code, Message string }
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if w.Code != 400 || response.Error.Code != "invalid_input" || response.Error.Message != "invalid draft" {
				t.Fatal(w.Code, response)
			}
		})
	}
}
