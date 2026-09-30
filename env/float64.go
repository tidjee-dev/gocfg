package env

import "strconv"

// Float64 parses the environment value for key as a float64,
// or returns fallback when unset.
func Float64(key string, fallback float64) (float64, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, parseError(key, "float", v)
	}
	return f, nil
}
