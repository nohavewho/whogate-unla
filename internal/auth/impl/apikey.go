package impl

import (
	"context"
	"fmt"
	"net/http"
)

// APIKeyAuthenticator implements API key authentication
type APIKeyAuthenticator struct {
	Header string
	ArgKey string
}

// Authenticate implements types.Authenticator.Authenticate
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, r *http.Request) error {
	// Get API key from header
	apiKey := r.Header.Get(a.Header)
	if apiKey == "" {
		return fmt.Errorf("missing %s header", a.Header)
	}

	// Store API key in context for later use
	ctx = context.WithValue(ctx, a.ArgKey, apiKey)
	*r = *r.WithContext(ctx)

	return nil
}
