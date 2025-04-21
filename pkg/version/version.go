package version

import (
	_ "embed"
)

//go:embed VERSION
var version string

// Get returns the current version of the application
func Get() string {
	return version
}
