package env

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Parser converts a raw environment value. It receives the variable key
// so custom parsers can build contextual errors (see *Error).
type Parser[T any] func(key, value string) (T, error)

// Var is an explicit, order-preserving configuration definition: the
// single source of truth the CLI consumes (see Any). Zero value is not
// usable; build Vars with the typed constructors.
type Var[T any] struct {
	Key      string
	Default  T
	Kind     string
	Secret   bool
	Required bool
	Parse    Parser[T]
}

// varCfg collects VarOpt functional options. It is deliberately not
// generic so call sites read plainly: env.StringVar("K", "", env.Secret()).
type varCfg struct {
	secret   bool
	required bool
}

// VarOpt configures a Var under construction.
type VarOpt func(*varCfg)

// Secret marks the Var's value as secret: it is redacted from error
// messages and left empty in generated `.env.example` files, even when
// the key does not match the secret heuristic (see IsSecretKey).
func Secret() VarOpt {
	return func(c *varCfg) { c.secret = true }
}

// Require marks the Var as required: an unset or empty value resolves
// to *RequiredError instead of the default. (Named Require because
// Required is already the required-value getter.)
func Require() VarOpt {
	return func(c *varCfg) { c.required = true }
}

func applyOpts(opts []VarOpt) varCfg {
	var c varCfg
	for _, o := range opts {
		o(&c)
	}
	return c
}

// StringVar defines a string variable.
func StringVar(key, def string, opts ...VarOpt) Var[string] {
	c := applyOpts(opts)
	return Var[string]{Key: key, Default: def, Kind: "string", Secret: c.secret, Required: c.required, Parse: parseString}
}

// BoolVar defines a boolean variable (see Bool for accepted spellings).
func BoolVar(key string, def bool, opts ...VarOpt) Var[bool] {
	c := applyOpts(opts)
	return Var[bool]{Key: key, Default: def, Kind: "bool", Secret: c.secret, Required: c.required, Parse: parseBool}
}

// IntVar defines an integer variable.
func IntVar(key string, def int, opts ...VarOpt) Var[int] {
	c := applyOpts(opts)
	return Var[int]{Key: key, Default: def, Kind: "int", Secret: c.secret, Required: c.required, Parse: parseInt}
}

// Int64Var defines an int64 variable.
func Int64Var(key string, def int64, opts ...VarOpt) Var[int64] {
	c := applyOpts(opts)
	return Var[int64]{Key: key, Default: def, Kind: "int64", Secret: c.secret, Required: c.required, Parse: parseInt64}
}

// Float64Var defines a float64 variable.
func Float64Var(key string, def float64, opts ...VarOpt) Var[float64] {
	c := applyOpts(opts)
	return Var[float64]{Key: key, Default: def, Kind: "float64", Secret: c.secret, Required: c.required, Parse: parseFloat64}
}

// DurationVar defines a duration variable (time.ParseDuration semantics).
func DurationVar(key string, def time.Duration, opts ...VarOpt) Var[time.Duration] {
	c := applyOpts(opts)
	return Var[time.Duration]{Key: key, Default: def, Kind: "duration", Secret: c.secret, Required: c.required, Parse: parseDuration}
}

// Resolve returns the Var's value following OS > .env-loaded env > Default.
// An unset or empty value yields the Default, or *RequiredError when the
// Var Require()s it. Parse failures yield *Error, redacted when the Var
// is Secret or the key matches the secret heuristic. Custom Parse funcs
// returning *Error for the same key keep their error with redaction
// enforced; any other error from a secret Var is replaced by a redacted
// *Error so custom messages cannot leak secrets.
func (v Var[T]) Resolve() (T, error) {
	var zero T
	secret := v.Secret || IsSecretKey(v.Key)
	raw, ok := lookup(v.Key)
	if !ok || raw == "" {
		if v.Required {
			return zero, &RequiredError{Key: v.Key}
		}
		return v.Default, nil
	}
	got, err := v.Parse(v.Key, raw)
	if err == nil {
		return got, nil
	}
	var perr *Error
	if errors.As(err, &perr) && perr.Key == v.Key {
		if secret {
			perr.Secret = true
			perr.Value = ""
		}
		return zero, perr
	}
	if secret {
		return zero, &Error{Key: v.Key, Kind: v.Kind, Secret: true}
	}
	return zero, err
}

// Any is the type-erased view of a Var for CLI consumption
// (`.env` generation, validation). Apps expose Definitions() []Any.
type Any interface {
	AnyKey() string
	AnyDefault() string
	AnyKind() string
	IsSecret() bool
	IsRequired() bool
}

// AnyKey reports the variable key.
func (v Var[T]) AnyKey() string { return v.Key }

// AnyDefault renders the default for generated files. Check IsRequired
// first: required Vars carry an ignored zero default. Rendering is
// canonical: Stringer types (URL, IP) use String, slices join with
// commas, anything else uses %v.
func (v Var[T]) AnyDefault() string {
	d := any(v.Default)
	if d == nil {
		return ""
	}
	// Guard typed nils (e.g. a *url.URL default that failed to parse:
	// calling String on it would panic).
	if rv := reflect.ValueOf(d); rv.Kind() == reflect.Ptr && rv.IsNil() {
		return ""
	}
	switch d := d.(type) {
	case nil:
		return ""
	case fmt.Stringer:
		return d.String()
	case []string:
		return strings.Join(d, ",")
	case []bool:
		strs := make([]string, 0, len(d))
		for _, b := range d {
			strs = append(strs, fmt.Sprintf("%v", b))
		}
		return strings.Join(strs, ",")
	default:
		return fmt.Sprintf("%v", v.Default)
	}
}

// AnyKind reports the type name ("string", "bool", "int",
// "int64", "float64", "duration", "url", "ip",
// "stringslice", "boolslice").
func (v Var[T]) AnyKind() string { return v.Kind }

// IsSecret reports whether the value must be redacted and left empty
// in generated examples (explicit Secret option or key heuristic).
func (v Var[T]) IsSecret() bool { return v.Secret || IsSecretKey(v.Key) }

// IsRequired reports whether an unset or empty value is an error.
func (v Var[T]) IsRequired() bool { return v.Required }
