package env

// String returns the environment value for key, or fallback when unset.
func String(key, fallback string) (string, error) {
	if v, ok := lookup(key); ok {
		return v, nil
	}
	return fallback, nil
}
