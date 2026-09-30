package env

import "strconv"

// Float64 parses the environment value for key as a float64,
// or returns fallback when unset.
func Float64(key string, fallback float64) (float64, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseFloat64(key, v)
}

// parseFloat64 is the shared parser used by Float64 and Float64Var.
func parseFloat64(key, value string) (float64, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, parseError(key, "float", value)
	}
	return f, nil
}
