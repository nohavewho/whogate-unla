package dto

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
}

// InitializeRequest represents an initialization request
type InitializeRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
