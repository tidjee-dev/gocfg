package env

// Required returns the environment value for key.
// Unset or empty is an error: *RequiredError.
func Required(key string) (string, error) {
	if v, ok := lookup(key); ok && v != "" {
		return v, nil
	}
	return "", &RequiredError{Key: key}
}
