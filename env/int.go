package env

import "strconv"

// Int parses the environment value for key as an int,
// or returns fallback when unset.
func Int(key string, fallback int) (int, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseInt(key, v)
}

// parseInt is the shared integer parser used by Int and IntVar.
func parseInt(key, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, parseError(key, "integer", value)
	}
	return n, nil
}
