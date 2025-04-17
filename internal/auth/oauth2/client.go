package oauth2

import (
	"context"
)

// Claims represents the JWT claims
type Claims map[string]any

// Client defines the interface for OAuth2 client
type Client interface {
	// ValidateToken validates the token and returns the claims
	ValidateToken(ctx context.Context, token string) (Claims, error)
}

// WithClaims stores claims in context
func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

// GetClaims retrieves claims from context
func GetClaims(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(Claims)
	return claims, ok
}

type claimsKey struct{}
