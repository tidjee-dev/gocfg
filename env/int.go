package env

import "strconv"

// Int parses the environment value for key as an int,
// or returns fallback when unset.
func Int(key string, fallback int) (int, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, parseError(key, "integer", v)
	}
	return n, nil
}
