package types

import (
	"context"
	"net/http"
)

// Authenticator defines the interface for authentication
type Authenticator interface {
	// Authenticate authenticates the request
	Authenticate(ctx context.Context, r *http.Request) error
}

// Mode represents the authentication mode
type Mode string

const (
	// ModeNone represents no authentication
	ModeNone Mode = "none"
	// ModeBearer represents bearer token authentication
	ModeBearer Mode = "bearer"
	// ModeAPIKey represents API key authentication
	ModeAPIKey Mode = "apikey"
)
