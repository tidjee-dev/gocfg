// Package env provides type-safe, error-returning environment lookups.
//
// Values resolve as OS environment > .env-loaded environment > fallback.
// Call gocfg.LoadEnv first to populate the process environment from .env
// files; this package only reads via os.LookupEnv, so OS values always win.
//
// Every getter returns (T, error). Invalid values are errors, never silent
// fallbacks. Error messages redact values for secret keys.
package env

import "os"

// lookup returns the raw value and whether the key is set.
// Unset and empty are distinguished: empty string with ok=true means
// the variable is set but empty.
func lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}
