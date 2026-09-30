package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckGreen(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeNamedEnv(t, dir, ".env", "GOCFG_C_ONE=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_C_ONE=\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"check"})
	if err := root.Execute(); err != nil {
		t.Fatalf("err: %v\nout:\n%s", err, sb.String())
	}
	for _, want := range []string{
		"✓ configuration directory",
		"✓ environment file",
		"✓ environment example",
		"✓ keys present",
		"Configuration looks consistent.",
	} {
		if !strings.Contains(sb.String(), want) {
			t.Fatalf("missing %q in:\n%s", want, sb.String())
		}
	}
}

func TestCheckMissingConfigDir(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_CD_ONE=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_CD_ONE=\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"check"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(sb.String(), "✗ configuration directory") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
}

func TestCheckMissingKeyFails(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeNamedEnv(t, dir, ".env", "GOCFG_CK_ONE=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_CK_ONE=\nGOCFG_CK_TWO=\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"check"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(sb.String(), "✗ key GOCFG_CK_TWO") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
}

func TestCheckOSSatisfiesWithoutEnvFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeNamedEnv(t, dir, ".env.example", "GOCFG_CO_ONLY=\n", 0o644)
	t.Setenv("GOCFG_CO_ONLY", "from-os")
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"check"})
	if err := root.Execute(); err != nil {
		t.Fatalf("OS-satisfied keys must pass: %v\nout:\n%s", err, sb.String())
	}
	if !strings.Contains(sb.String(), ".env not found") {
		t.Fatalf("expected missing-file warning in:\n%s", sb.String())
	}
}

func TestCheckEmptyCountsAsPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Empty in .env: check passes (presence), validate would fail (values).
	writeNamedEnv(t, dir, ".env", "GOCFG_CE_EMPTY=\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_CE_EMPTY=\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"check"})
	if err := root.Execute(); err != nil {
		t.Fatalf("empty-but-present must pass check: %v\nout:\n%s", err, sb.String())
	}
}

func TestCheckMalformedEnvFails(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeNamedEnv(t, dir, ".env", "BROKEN LINE\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_CM_ONE=\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"check"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(sb.String(), "✗ environment file") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
}

func TestCheckExtrasWarnOnly(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeNamedEnv(t, dir, ".env", "GOCFG_CX_ONE=1\nGOCFG_CX_EXTRA=9\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_CX_ONE=\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"check"})
	if err := root.Execute(); err != nil {
		t.Fatalf("extras must not fail: %v\nout:\n%s", err, sb.String())
	}
	if !strings.Contains(sb.String(), "! GOCFG_CX_EXTRA (not in .env.example)") {
		t.Fatalf("missing extra warning in:\n%s", sb.String())
	}
}
