package env

import "strconv"

// Int64 parses the environment value for key as an int64,
// or returns fallback when unset.
func Int64(key string, fallback int64) (int64, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseInt64(key, v)
}

// parseInt64 is the shared parser used by Int64 and Int64Var.
func parseInt64(key, value string) (int64, error) {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, parseError(key, "integer", value)
	}
	return n, nil
}
