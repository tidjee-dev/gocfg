// Package gocfg is a small, type-safe configuration library for Go.
//
// Configuration is written in Go (see the env package), environment-specific
// values come from .env files and OS environment variables, and the
// application composes its own Config explicitly.
package gocfg

import "github.com/tidjee-dev/gocfg/internal/dotenv"

// LoadEnv loads paths (default ".env") into the process environment.
// Keys already present in the OS environment are never overridden,
// enforcing OS > .env > Go default. A missing default ".env" is a no-op.
func LoadEnv(paths ...string) error {
	return dotenv.Load(paths...)
}
