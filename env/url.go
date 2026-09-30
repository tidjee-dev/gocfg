package env

import "net/url"

// URL parses the environment value for key as a URL with a scheme
// (e.g. "https://example.com", "postgres://localhost/db"),
// or returns fallback when unset.
func URL(key, fallback string) (*url.URL, error) {
	v, ok := lookup(key)
	if !ok {
		return parseURL(key, fallback)
	}
	return parseURL(key, v)
}

// parseURL is the shared parser used by URL and URLVar.
func parseURL(key, value string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil || u.Scheme == "" {
		return nil, parseError(key, "url", value)
	}
	return u, nil
}

// URLVar defines a URL variable. An invalid default yields a nil URL;
// prefer a valid default or Require().
func URLVar(key, def string, opts ...VarOpt) Var[*url.URL] {
	d, _ := parseURL(key, def)
	c := applyOpts(opts)
	return Var[*url.URL]{Key: key, Default: d, Kind: "url", Secret: c.secret, Required: c.required, Parse: parseURL}
}
