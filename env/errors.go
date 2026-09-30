package env

import (
	"fmt"
	"strings"
)

// Error is a contextual configuration error. Values of secret keys are
// never included in the message.
type Error struct {
	Key    string // environment variable name
	Kind   string // expected type, e.g. "integer", "boolean"
	Value  string // raw value; empty when Secret
	Secret bool   // value redacted
}

func (e *Error) Error() string {
	if e.Secret {
		return fmt.Sprintf("invalid value for %s: expected %s", e.Key, e.Kind)
	}
	return fmt.Sprintf("invalid value for %s: %q: expected %s", e.Key, e.Value, e.Kind)
}

// RequiredError reports a missing required variable.
type RequiredError struct {
	Key string
}

func (e *RequiredError) Error() string {
	return fmt.Sprintf("required environment variable %s is not set", e.Key)
}

// secretMarkers backs the v0.1 heuristic. v0.2 adds an explicit
// Secret flag on Var[T]; both trigger redaction.
var secretMarkers = []string{"PASSWORD", "SECRET", "KEY", "TOKEN"}

// IsSecretKey reports whether key looks like a secret (case-insensitive
// substring match). Use it to decide redaction.
func IsSecretKey(key string) bool {
	upper := strings.ToUpper(key)
	for _, m := range secretMarkers {
		if strings.Contains(upper, m) {
			return true
		}
	}
	return false
}

// parseError builds a redacting *Error for key.
func parseError(key, kind, value string) *Error {
	return parseErrorSecret(key, kind, value, false)
}

// parseErrorSecret builds a redacting *Error for key, forcing redaction
// when secret is true (e.g. a Var with the Secret option) even if the
// key does not look secret.
func parseErrorSecret(key, kind, value string, secret bool) *Error {
	if secret || IsSecretKey(key) {
		return &Error{Key: key, Kind: kind, Secret: true}
	}
	return &Error{Key: key, Kind: kind, Value: value}
}
