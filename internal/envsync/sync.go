// Package envsync synchronizes `.env` from `.env.example`
// (append-missing-only). Existing values, comments, blank lines and
// order are never touched; writes are atomic. It is stdlib-only core:
// the cobra wiring lives in internal/cli.
package envsync

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tidjee-dev/gocfg/env"
	"github.com/tidjee-dev/gocfg/internal/dotenv"
)

// Options controls a sync run.
type Options struct {
	// EnvFile is the file to complete (default ".env").
	EnvFile string
	// ExampleFile is the read-only schema (default ".env.example").
	ExampleFile string
	// Check reports what would change without writing.
	Check bool
}

// Result describes a sync run.
type Result struct {
	// Created reports the env file did not exist (or would be created in Check).
	Created bool
	// Added lists appended keys (or keys that would be appended in Check), sorted.
	Added []string
	// Warnings lists extras and secret values found in the example, sorted.
	Warnings []string
	// Changed reports Created || len(Added) > 0.
	Changed bool
}

// Run completes the env file from the example file.
func Run(opts Options) (*Result, error) {
	if opts.EnvFile == "" {
		opts.EnvFile = ".env"
	}
	if opts.ExampleFile == "" {
		opts.ExampleFile = ".env.example"
	}

	rawExample, err := os.ReadFile(opts.ExampleFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("envsync: no schema %s (run gocfg init)", opts.ExampleFile)
		}
		return nil, fmt.Errorf("envsync: %w", err)
	}
	schema, err := dotenv.Parse(rawExample)
	if err != nil {
		return nil, fmt.Errorf("envsync: invalid schema %s: %w", opts.ExampleFile, err)
	}

	rawEnv, err := os.ReadFile(opts.EnvFile)
	exists := true
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("envsync: %w", err)
		}
		exists = false
		rawEnv = nil
	}
	// Preserve the existing file mode across the atomic rename below.
	mode := os.FileMode(0o600)
	if exists {
		fi, err := os.Stat(opts.EnvFile)
		if err != nil {
			return nil, fmt.Errorf("envsync: %w", err)
		}
		mode = fi.Mode().Perm()
	}
	var present map[string]string
	if exists {
		present, err = dotenv.Parse(rawEnv)
		if err != nil {
			return nil, fmt.Errorf("envsync: invalid %s: %w", opts.EnvFile, err)
		}
	} else {
		present = map[string]string{}
	}

	res := &Result{Created: !exists}
	var lines []string
	for _, k := range slices.Sorted(maps.Keys(schema)) {
		if _, ok := present[k]; ok {
			continue
		}
		v := schema[k]
		if v != "" && env.IsSecretKey(k) {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("%s has a value in %s (secret left empty)", k, opts.ExampleFile))
			v = ""
		}
		lines = append(lines, k+"="+v)
		res.Added = append(res.Added, k)
	}
	for _, k := range slices.Sorted(maps.Keys(present)) {
		if _, ok := schema[k]; !ok {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("%s (not in %s)", k, opts.ExampleFile))
		}
	}
	slices.Sort(res.Warnings)
	res.Changed = res.Created || len(res.Added) > 0

	if opts.Check || !res.Changed {
		return res, nil
	}
	if err := writeMerged(opts.EnvFile, rawEnv, exists, mode, lines); err != nil {
		return nil, err
	}
	return res, nil
}

// writeMerged appends lines to raw (or creates the file), atomically.
// The replacement carries mode so existing permissions survive the rename.
func writeMerged(path string, raw []byte, exists bool, mode os.FileMode, lines []string) error {
	var out []byte
	switch {
	case len(lines) == 0:
		out = append([]byte{}, raw...)
	case !exists || len(raw) == 0:
		out = []byte(strings.Join(lines, "\n") + "\n")
	default:
		out = append([]byte{}, raw...)
		if out[len(out)-1] != '\n' {
			out = append(out, '\n')
		}
		out = append(out, []byte(strings.Join(lines, "\n")+"\n")...)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("envsync: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".gocfg-*")
	if err != nil {
		return fmt.Errorf("envsync: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		return fmt.Errorf("envsync: %w", err)
	}
	// Carry the mode over so existing permissions survive the rename
	// (new files default to 0600 via mode).
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return fmt.Errorf("envsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("envsync: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("envsync: %w", err)
	}
	return nil
}
