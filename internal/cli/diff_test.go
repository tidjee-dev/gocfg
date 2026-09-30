package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/tidjee-dev/gocfg/env"
)

func TestDiffClean(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_DF_A=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_DF_A=1\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"diff"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), "No differences.") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
}

func TestDiffFindsMissing(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_DG_A=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_DG_A=1\nGOCFG_DG_B=2\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"diff"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(sb.String(), "✗ GOCFG_DG_B: missing in .env") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
}

func TestDiffWithDefsDrift(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_DH_A=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_DH_A=1\n", 0o644)
	m := writeManifest(t, dir, []env.Any{env.StringVar("GOCFG_DH_A", "changed")})
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"diff", "--defs", m})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(sb.String(), "default drift") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
}
