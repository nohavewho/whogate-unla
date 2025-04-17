package impl

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcp-ecosystem/mcp-gateway/internal/auth/oauth2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOAuth2Client struct {
	mock.Mock
}

func (m *mockOAuth2Client) ValidateToken(ctx context.Context, token string) (oauth2.Claims, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(oauth2.Claims), args.Error(1)
}

func TestOAuth2_Authenticate(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		token         string
		claims        oauth2.Claims
		validateError error
		wantError     bool
	}{
		{
			name:       "valid token",
			authHeader: "Bearer valid-token",
			token:      "valid-token",
			claims:     oauth2.Claims{"sub": "user123"},
			wantError:  false,
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			wantError:  true,
		},
		{
			name:       "invalid authorization header format",
			authHeader: "InvalidFormat",
			wantError:  true,
		},
		{
			name:       "empty token",
			authHeader: "Bearer ",
			wantError:  true,
		},
		{
			name:          "invalid token",
			authHeader:    "Bearer invalid-token",
			token:         "invalid-token",
			validateError: assert.AnError,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(mockOAuth2Client)
			if tt.token != "" {
				mockClient.On("ValidateToken", mock.Anything, tt.token).Return(tt.claims, tt.validateError)
			}

			// Create authenticator
			authenticator := NewOAuth2(mockClient)

			// Create request
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Authenticate
			ctx, err := authenticator.Authenticate(context.Background(), req)

			// Check error
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Check claims in context
				claims, _ := oauth2.GetClaims(ctx)
				assert.Equal(t, tt.claims, claims)
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}
