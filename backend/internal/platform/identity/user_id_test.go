package identity

import (
	"context"
	"testing"
)

func TestUserID(t *testing.T) {
	ctx := context.Background()
	if UserID(ctx) != "" || UserID(WithUser(ctx, User{ID: "actor"})) != "actor" {
		t.Fatal("unexpected user ID")
	}
}
