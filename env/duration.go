package env

import "time"

// Duration parses the environment value for key with time.ParseDuration
// semantics (e.g. "30s", "1m30s"), or returns fallback when unset.
func Duration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseDuration(key, v)
}

// parseDuration is the shared parser used by Duration and DurationVar.
func parseDuration(key, value string) (time.Duration, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, parseError(key, "duration", value)
	}
	return d, nil
}
