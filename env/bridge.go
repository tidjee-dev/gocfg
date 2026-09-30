package env

import (
	"encoding/json"
	"fmt"
)

// ManifestVersion is the envelope version written by MarshalDefinitions.
// UnmarshalDefinitions rejects anything else so format evolution stays
// explicit (post-v1: accept older versions instead of breaking).
const ManifestVersion = 1

// Definition is the serializable form of one Var: the unit a manifest
// file carries from application code to the CLI. Default is the
// AnyDefault rendering; required Vars marshal with an empty Default
// since they must be filled in.
type Definition struct {
	Key      string `json:"key"`
	Kind     string `json:"kind"`
	Default  string `json:"default,omitempty"`
	Secret   bool   `json:"secret,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type manifest struct {
	Version int          `json:"version"`
	Vars    []Definition `json:"vars"`
}

// MarshalDefinitions encodes defs as an indented manifest document.
// A nil or empty input yields a manifest with no vars (still versioned).
func MarshalDefinitions(defs []Any) ([]byte, error) {
	m := manifest{Version: ManifestVersion}
	for _, d := range defs {
		def := Definition{
			Key:      d.AnyKey(),
			Kind:     d.AnyKind(),
			Secret:   d.IsSecret(),
			Required: d.IsRequired(),
		}
		if !def.Required {
			def.Default = d.AnyDefault()
		}
		m.Vars = append(m.Vars, def)
	}
	if m.Vars == nil {
		m.Vars = []Definition{}
	}
	return json.MarshalIndent(m, "", "  ")
}

// UnmarshalDefinitions decodes a manifest document. It checks the
// envelope version and that every key is non-empty; kind support is
// validated by consumers (unknown kinds fail at use, naming the key).
func UnmarshalDefinitions(data []byte) ([]Definition, error) {
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("env: invalid manifest: %w", err)
	}
	if m.Version != ManifestVersion {
		return nil, fmt.Errorf("env: unsupported manifest version %d (want %d)", m.Version, ManifestVersion)
	}
	for i, d := range m.Vars {
		if d.Key == "" {
			return nil, fmt.Errorf("env: manifest var %d has an empty key", i)
		}
	}
	return m.Vars, nil
}
