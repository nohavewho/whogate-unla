package impl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mcp-ecosystem/mcp-gateway/internal/auth/oauth2"
)

var (
	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")
)

// OAuth2 implements the auth.Authenticator interface using OAuth2
type OAuth2 struct {
	client oauth2.Client
}

// NewOAuth2 creates a new OAuth2 authenticator
func NewOAuth2(client oauth2.Client) *OAuth2 {
	return &OAuth2{
		client: client,
	}
}

// Authenticate implements the auth.Authenticator interface
func (o *OAuth2) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ctx, ErrUnauthorized
	}

	// Check if it's a Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ctx, ErrUnauthorized
	}

	token := parts[1]
	if token == "" {
		return ctx, ErrUnauthorized
	}

	// Validate token and get claims
	claims, err := o.client.ValidateToken(ctx, token)
	if err != nil {
		return ctx, fmt.Errorf("validate token: %w", err)
	}

	// Store claims in context
	return oauth2.WithClaims(ctx, claims), nil
}
