package env

import "strconv"

// Int64 parses the environment value for key as an int64,
// or returns fallback when unset.
func Int64(key string, fallback int64) (int64, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, parseError(key, "integer", v)
	}
	return n, nil
}
