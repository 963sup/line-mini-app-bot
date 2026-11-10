package google

import (
	"context"

	"github.com/963sup/line-mini-app-bot/internal/identity"
	"google.golang.org/api/idtoken"
)

type Verifier struct{ Audience, Email string }

func (v *Verifier) Verify(ctx context.Context, token string) error {
	claims, err := idtoken.Validate(ctx, token, v.Audience)
	if err != nil || claims == nil || claims.Claims["email"] != v.Email || claims.Claims["email_verified"] != true {
		return identity.ErrRejectedToken
	}
	return nil
}
