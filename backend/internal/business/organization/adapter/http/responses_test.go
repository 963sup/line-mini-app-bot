package organizationhttp

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	dto "github.com/963sup/line-mini-app-bot/internal/business/organization/application/dto"
)

func TestContextResponseContract(t *testing.T) {
	enterprise, detail := "enterprise", "org"
	for _, tc := range []struct {
		name    string
		context dto.Context
		json    string
	}{
		{"null values", dto.Context{}, `{"enterprises":null,"canCreateEnterprise":false,"groups":null,"selected":null,"canSwitch":false}`},
		{"empty arrays", dto.Context{Enterprises: []dto.EnterpriseLink{}, Groups: []dto.Group{}}, `{"enterprises":[],"canCreateEnterprise":false,"groups":[],"selected":null,"canSwitch":false}`},
		{"selected", dto.Context{
			Enterprises: []dto.EnterpriseLink{{ID: enterprise, Name: "HQ"}}, CanCreateEnterprise: true, CanSwitch: true,
			Groups:   []dto.Group{{EnterpriseID: &enterprise, Label: "HQ", Organizations: []dto.Option{{ID: "org", Name: "Taipei", Detail: &detail}}}},
			Selected: &dto.Selected{ID: "org", Name: "Taipei", GroupLabel: "HQ"},
		}, `{"enterprises":[{"id":"enterprise","name":"HQ"}],"canCreateEnterprise":true,"groups":[{"enterpriseId":"enterprise","label":"HQ","organizations":[{"organizationId":"org","name":"Taipei","detail":"org"}]}],"selected":{"organizationId":"org","name":"Taipei","groupLabel":"HQ"},"canSwitch":true}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := contextResponse(&tc.context, nil)
			if err != nil {
				t.Fatal(err)
			}
			body, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.json), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("response changed: %s", body)
			}
		})
	}
	denied := errors.New("denied")
	if value, err := contextResponse(&dto.Context{}, denied); value != nil || !errors.Is(err, denied) {
		t.Fatal("error response exposed context", value, err)
	}
	if value, err := contextResponse(nil, nil); value != nil || err != nil {
		t.Fatal(value, err)
	}
}
