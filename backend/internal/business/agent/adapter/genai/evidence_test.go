package genai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type generator func(context.Context, string, string) (string, error)

func (g generator) Generate(ctx context.Context, instruction, source string) (string, error) {
	return g(ctx, instruction, source)
}

func TestQuoteResponseContract(t *testing.T) {
	providerFailure := errors.New("provider failure")
	for _, tc := range []struct {
		name, response, want string
		providerError        error
		absent, invalid      bool
	}{
		{name: "literal", response: `"數量 20 SET"`, want: "數量 20 SET"},
		{name: "escaped", response: `"數量\n20 SET"`, want: "數量\n20 SET"},
		{name: "absent", response: "null", absent: true},
		{name: "empty string", response: `""`},
		{name: "plain text", response: "20 SET", invalid: true},
		{name: "object", response: `{"quote":"20 SET"}`, invalid: true},
		{name: "multiple values", response: `"20 SET" null`, invalid: true},
		{name: "provider unavailable", providerError: providerFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			type requestKey struct{}
			key := requestKey{}
			ctx := context.WithValue(t.Context(), key, "request")
			m := EvidenceModel{Generator: generator(func(got context.Context, instruction, source string) (string, error) {
				if got.Value(key) != "request" || source != "page text" || !strings.Contains(instruction, "JSON string") || !strings.Contains(instruction, "untrusted data") {
					t.Fatal("generation request changed", source, instruction)
				}
				return tc.response, tc.providerError
			})}
			quote, err := m.Quote(ctx, "page text")
			if tc.providerError != nil {
				if !errors.Is(err, tc.providerError) || quote != nil {
					t.Fatal(quote, err)
				}
				return
			}
			if tc.invalid {
				if err == nil || quote != nil {
					t.Fatal("invalid provider format accepted", quote, err)
				}
				return
			}
			if err != nil || (quote == nil) != tc.absent {
				t.Fatal(quote, err)
			}
			if quote != nil && *quote != tc.want {
				t.Fatal(*quote)
			}
		})
	}
}
