package cli

import (
	"strings"
	"testing"
)

func TestVersionOutputsVersion(t *testing.T) {
	Version = "v0.1.0-test"
	t.Cleanup(func() { Version = "dev" })

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := sb.String()
	if !strings.Contains(out, "gocfg") || !strings.Contains(out, "v0.1.0-test") {
		t.Fatalf("unexpected version output: %q", out)
	}
}

func TestHelpSucceeds(t *testing.T) {
	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), "gocfg") {
		t.Fatalf("unexpected help output: %q", sb.String())
	}
}
