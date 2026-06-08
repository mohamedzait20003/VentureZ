package auth

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	
	return false
}

type ctxKey struct{}

func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok
}
