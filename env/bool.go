package env

import "strings"

// Accepted boolean spellings (case-insensitive, trimmed).
var (
	trueSet  = map[string]bool{"1": true, "true": true, "yes": true, "y": true, "on": true}
	falseSet = map[string]bool{"0": true, "false": true, "no": true, "n": true, "off": true}
)

// Bool parses the environment value for key as a boolean,
// or returns fallback when unset.
func Bool(key string, fallback bool) (bool, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseBool(key, v)
}

// parseBool is the shared boolean parser used by Bool and BoolVar.
func parseBool(key, value string) (bool, error) {
	norm := strings.ToLower(strings.TrimSpace(value))
	if trueSet[norm] {
		return true, nil
	}
	if falseSet[norm] {
		return false, nil
	}
	return false, parseError(key, "boolean", value)
}
