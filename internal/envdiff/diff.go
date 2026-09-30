// Package envdiff compares configuration sources file-level: an
// optional definitions manifest against `.env.example` and `.env`.
// It reports structure (which keys where, default drift), never values:
// secrets render as such and non-secret values stay out (see export).
// OS environment is deliberately ignored: diff answers "are the files
// in sync", check answers "will the app resolve".
package envdiff

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/tidjee-dev/gocfg/env"
	"github.com/tidjee-dev/gocfg/internal/dotenv"
)

// Status classifies one row: Ok, Fail (missing/drift, exits 1) or
// Warn (extras, secret values, example-only keys).
type Status int

const (
	Ok Status = iota
	Fail
	Warn
)

// Row is one key's comparison result. Detail carries any non-secret
// defaults involved; secret material never appears.
type Row struct {
	Key    string
	Status Status
	Detail string // human note, e.g. "missing in .env"
	Secret bool
}

// Compare loads the files (and optional manifest) and returns one row
// per key in the union, sorted. A nil manifest means two-way comparison.
func Compare(envFile, exampleFile string, defs []env.Definition) ([]Row, error) {
	rawExample, err := os.ReadFile(exampleFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("envdiff: no schema %s (run gocfg init)", exampleFile)
		}
		return nil, fmt.Errorf("envdiff: %w", err)
	}
	example, err := dotenv.Parse(rawExample)
	if err != nil {
		return nil, fmt.Errorf("envdiff: invalid schema %s: %w", exampleFile, err)
	}
	present := map[string]string{}
	if rawEnv, err := os.ReadFile(envFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("envdiff: %w", err)
		}
	} else if present, err = dotenv.Parse(rawEnv); err != nil {
		return nil, fmt.Errorf("envdiff: invalid %s: %w", envFile, err)
	}

	manifest := map[string]env.Definition{}
	for _, d := range defs {
		manifest[d.Key] = d
	}
	keys := map[string]bool{}
	for k := range example {
		keys[k] = true
	}
	for k := range present {
		keys[k] = true
	}
	for k := range manifest {
		keys[k] = true
	}

	var rows []Row
	for _, k := range slices.Sorted(maps.Keys(keys)) {
		rows = append(rows, classify(k, example, present, manifest, exampleFile))
	}
	return rows, nil
}

func secretOf(k string, manifest map[string]env.Definition) bool {
	if d, ok := manifest[k]; ok && d.Secret {
		return true
	}
	return env.IsSecretKey(k)
}

func classify(k string, example, present map[string]string, manifest map[string]env.Definition, exampleFile string) Row {
	secret := secretOf(k, manifest)
	_, inExample := example[k]
	_, inEnv := present[k]
	d, inManifest := manifest[k]

	switch {
	case !inExample && !inEnv && inManifest:
		return Row{Key: k, Status: Fail, Detail: fmt.Sprintf("in manifest but not in %s nor .env", exampleFile), Secret: secret}
	case inManifest && inExample && !secret && d.Default != example[k]:
		return Row{Key: k, Status: Fail, Secret: secret,
			Detail: fmt.Sprintf("default drift: manifest=%q example=%q", d.Default, example[k])}
	case inExample && !inEnv:
		return Row{Key: k, Status: Fail, Detail: "missing in .env"}
	case !inExample && inEnv:
		return Row{Key: k, Status: Warn, Detail: fmt.Sprintf("extra (not in %s)", exampleFile)}
	case inManifest && !inExample:
		return Row{Key: k, Status: Warn, Secret: secret, Detail: "in manifest but not in example"}
	}
	// Present everywhere expected: surface secret values in the example.
	if inExample && secret && example[k] != "" {
		return Row{Key: k, Status: Warn, Secret: true,
			Detail: fmt.Sprintf("has a value in %s", exampleFile)}
	}
	return Row{Key: k, Status: Ok, Secret: secret, Detail: "consistent"}
}
