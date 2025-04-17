package errors

import "fmt"

// ErrDuplicateToolName is returned when a tool name is duplicated
func ErrDuplicateToolName(name string) error {
	return fmt.Errorf("duplicate tool name: %s", name)
}

// ErrDuplicateServerName is returned when a server name is duplicated
func ErrDuplicateServerName(name string) error {
	return fmt.Errorf("duplicate server name: %s", name)
}

// ErrDuplicateRouterPrefix is returned when a router prefix is duplicated
func ErrDuplicateRouterPrefix(prefix string) error {
	return fmt.Errorf("duplicate router prefix: %s", prefix)
}

// ErrInvalidAuthMode is returned when an invalid authentication mode is specified
func ErrInvalidAuthMode(mode string) error {
	return fmt.Errorf("invalid auth mode: %s", mode)
}

// ErrToolNotFound is returned when a tool is not found
func ErrToolNotFound(name string) error {
	return fmt.Errorf("tool not found: %s", name)
}

// ErrServerNotFound is returned when a server is not found
func ErrServerNotFound(name string) error {
	return fmt.Errorf("server not found: %s", name)
}
