package impl

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/mcp-ecosystem/mcp-gateway/internal/auth/types"
)

// BasicAuthenticator implements Basic authentication
type BasicAuthenticator struct {
	Realm string
}

// Authenticate implements types.Authenticator.Authenticate
func (a *BasicAuthenticator) Authenticate(ctx context.Context, r *http.Request) error {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("missing Authorization header")
	}

	if !strings.HasPrefix(auth, "Basic ") {
		return fmt.Errorf("invalid Authorization header format")
	}

	// Decode credentials
	credentials, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %v", err)
	}

	// Split username and password
	parts := strings.SplitN(string(credentials), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid credentials format")
	}

	username, password := parts[0], parts[1]

	// Store credentials in context for later use
	ctx = context.WithValue(ctx, types.ContextKeyUsername, username)
	ctx = context.WithValue(ctx, types.ContextKeyPassword, password)
	*r = *r.WithContext(ctx)

	return nil
}
