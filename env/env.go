// Package env provides type-safe, error-returning environment lookups.
//
// Values resolve as OS environment > .env-loaded environment > fallback.
// Call gocfg.LoadEnv first to populate the process environment from .env
// files; this package only reads via os.LookupEnv, so OS values always win.
//
// Every getter returns (T, error). Invalid values are errors, never silent
// fallbacks. Error messages redact values for secret keys.
//
// Explicit definitions live in Var: a Var[T] bundles key, default, type,
// required and secret metadata with a shared parser, and resolves with
// the same precedence. Apps expose Definitions() []Any for CLI consumption.
//
// Naming note: Required is the required-value getter; Require() is the
// Var option marking a definition required. The pair is intentional.
package env

import "os"

// lookup returns the raw value and whether the key is set.
// Unset and empty are distinguished: empty string with ok=true means
// the variable is set but empty.
func lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}
