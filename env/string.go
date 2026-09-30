package env

// String returns the environment value for key, or fallback when unset.
func String(key, fallback string) (string, error) {
	if v, ok := lookup(key); ok {
		return v, nil
	}
	return fallback, nil
}

// parseString is the shared parser used by String and StringVar.
// Strings never fail to parse.
func parseString(_, value string) (string, error) {
	return value, nil
}
