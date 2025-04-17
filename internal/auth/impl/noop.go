package impl

import (
	"context"
	"net/http"
)

// NoopAuthenticator implements no-op authentication
type NoopAuthenticator struct {
	Header string
	ArgKey string
}

// Authenticate implements Authenticator.Authenticate
func (a *NoopAuthenticator) Authenticate(ctx context.Context, r *http.Request) error {
	// No-op authentication
	return nil
}
