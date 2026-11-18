package firestorekey

import "testing"

func TestCursorCompatibility(t *testing.T) {
	const id = "0123456789abcdef0123456789abcdef"
	const token = "eyJTY29wZSI6IjE5NWEwNThkNjI4YWJjMmY3NDlhMDdhZWZkZmQ3Y2Y1NzUxMDM3YzRlYzQ3YTVmM2QyMGE0NzJjNTU4NjE2NDIiLCJMYXN0IjoiMDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWYifQ"
	scope := []string{"owner", "source"}
	if got := EncodeCursor(scope, id); got != token {
		t.Fatalf("cursor changed: %s", got)
	}
	if last, ok := DecodeCursor(token, scope); !ok || last != id {
		t.Fatal(last, ok)
	}
	if _, ok := DecodeCursor(token, []string{"other"}); ok {
		t.Fatal("accepted another scope")
	}
	if _, ok := DecodeCursor("!", scope); ok {
		t.Fatal("accepted malformed cursor")
	}
}
