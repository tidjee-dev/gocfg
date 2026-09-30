package env

import (
	"strings"
)

// splitList splits s on commas, trimming spaces and honoring
// single- and double-quoted segments (`"a,b", 'c,d', e` yields
// ["a,b", "c,d", "e"]; quotes group but are stripped).
// Empty segments are dropped.
func splitList(s string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if seg := strings.TrimSpace(cur.String()); seg != "" {
			out = append(out, seg)
		}
		cur.Reset()
	}
	var quote rune // 0 when outside quotes
	for _, r := range s {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote == 0 && (r == '"' || r == '\''):
			quote = r
		case r == ',' && quote == 0:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

// StringSlice parses the environment value for key as a comma-separated
// list (see splitList), or returns fallback when unset.
func StringSlice(key string, fallback []string) ([]string, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseStringSlice(key, v)
}

// parseStringSlice is the shared parser used by StringSlice and StringSliceVar.
func parseStringSlice(_, value string) ([]string, error) {
	return splitList(value), nil
}

// StringSliceVar defines a string-slice variable.
func StringSliceVar(key string, def []string, opts ...VarOpt) Var[[]string] {
	c := applyOpts(opts)
	return Var[[]string]{Key: key, Default: def, Kind: "stringslice", Secret: c.secret, Required: c.required, Parse: parseStringSlice}
}

// BoolSlice parses the environment value for key as a comma-separated
// list of booleans (see Bool for accepted spellings),
// or returns fallback when unset.
func BoolSlice(key string, fallback []bool) ([]bool, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseBoolSlice(key, v)
}

// parseBoolSlice is the shared parser used by BoolSlice and BoolSliceVar.
func parseBoolSlice(key, value string) ([]bool, error) {
	var out []bool
	for _, elem := range splitList(value) {
		b, err := parseBool(key, elem)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

// BoolSliceVar defines a boolean-slice variable.
func BoolSliceVar(key string, def []bool, opts ...VarOpt) Var[[]bool] {
	c := applyOpts(opts)
	return Var[[]bool]{Key: key, Default: def, Kind: "boolslice", Secret: c.secret, Required: c.required, Parse: parseBoolSlice}
}
