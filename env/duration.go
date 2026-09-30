package env

import "time"

// Duration parses the environment value for key with time.ParseDuration
// semantics (e.g. "30s", "1m30s"), or returns fallback when unset.
func Duration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, parseError(key, "duration", v)
	}
	return d, nil
}
