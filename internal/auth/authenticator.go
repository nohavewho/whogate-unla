package auth

import (
	"context"
	"net/http"

	"github.com/mcp-ecosystem/mcp-gateway/internal/auth/impl"
	"github.com/mcp-ecosystem/mcp-gateway/internal/auth/types"
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

// NewAuthenticator creates a new authenticator based on the mode
func NewAuthenticator(mode types.Mode, header, argKey string) types.Authenticator {
	switch mode {
	case types.ModeBearer:
		return &impl.BearerAuthenticator{
			Header: header,
			ArgKey: argKey,
		}
	case types.ModeAPIKey:
		return &impl.APIKeyAuthenticator{
			Header: header,
			ArgKey: argKey,
		}
	case types.ModeNone:
		return &impl.NoopAuthenticator{
			Header: header,
			ArgKey: argKey,
		}
	default:
		return &impl.NoopAuthenticator{
			Header: header,
			ArgKey: argKey,
		}
	}
}
