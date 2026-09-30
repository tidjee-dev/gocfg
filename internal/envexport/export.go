// Package envexport renders resolved configuration values for shells
// and tools. It is strictly schema-bounded: only keys declared in the
// example file are emitted, never the whole OS environment. It performs
// no process-environment mutation. Values print as-is on explicit user
// request (that is the point of export); keep the output out of logs.
package envexport

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/tidjee-dev/gocfg/internal/dotenv"
)

// Options controls a collection run.
type Options struct {
	// EnvFile is the file to read (default ".env", may be absent).
	EnvFile string
	// ExampleFile is the schema bounding output (default ".env.example").
	ExampleFile string
}

// Entry is one resolved variable in schema (sorted) order.
type Entry struct {
	Key     string
	Value   string
	Missing bool // absent from both the env file and OS
}

// Collect resolves every schema key as OS > env file.
func Collect(opts Options) ([]Entry, error) {
	if opts.EnvFile == "" {
		opts.EnvFile = ".env"
	}
	if opts.ExampleFile == "" {
		opts.ExampleFile = ".env.example"
	}
	rawExample, err := os.ReadFile(opts.ExampleFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("envexport: no schema %s (run gocfg init)", opts.ExampleFile)
		}
		return nil, fmt.Errorf("envexport: %w", err)
	}
	schema, err := dotenv.Parse(rawExample)
	if err != nil {
		return nil, fmt.Errorf("envexport: invalid schema %s: %w", opts.ExampleFile, err)
	}
	fileVars := map[string]string{}
	if rawEnv, err := os.ReadFile(opts.EnvFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("envexport: %w", err)
		}
	} else if fileVars, err = dotenv.Parse(rawEnv); err != nil {
		return nil, fmt.Errorf("envexport: invalid %s: %w", opts.EnvFile, err)
	}
	var out []Entry
	for _, k := range slices.Sorted(maps.Keys(schema)) {
		if v, ok := os.LookupEnv(k); ok {
			out = append(out, Entry{Key: k, Value: v})
			continue
		}
		if v, ok := fileVars[k]; ok {
			out = append(out, Entry{Key: k, Value: v})
			continue
		}
		out = append(out, Entry{Key: k, Missing: true})
	}
	return out, nil
}

// FormatShell renders `export KEY='value'` lines (POSIX single-quote escaping).
func FormatShell(entries []Entry) string {
	var sb strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&sb, "export %s='%s'\n", e.Key, strings.ReplaceAll(e.Value, `'`, `'\''`))
	}
	return sb.String()
}

// FormatDotenv renders `KEY=value` lines, double-quoting values that
// contain newlines, quotes or significant surrounding whitespace.
func FormatDotenv(entries []Entry) string {
	var sb strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&sb, "%s=%s\n", e.Key, dotenvEncode(e.Value))
	}
	return sb.String()
}

func dotenvEncode(v string) string {
	if strings.ContainsAny(v, "\n\r\"") || strings.TrimSpace(v) != v {
		r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`)
		return `"` + r.Replace(v) + `"`
	}
	return v
}

// FormatJSON renders {"KEY": "value"} (indented for humans, jq-safe).
func FormatJSON(entries []Entry) (string, error) {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("envexport: %w", err)
	}
	return string(raw) + "\n", nil
}
