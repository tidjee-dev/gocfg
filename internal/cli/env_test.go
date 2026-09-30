package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvSyncsMissing(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_E_ONE=keep\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_E_ONE=x\nGOCFG_E_TWO=new\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"env"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), "added GOCFG_E_TWO") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".env"))
	if string(b) != "GOCFG_E_ONE=keep\nGOCFG_E_TWO=new\n" {
		t.Fatalf("unexpected .env: %q", b)
	}
}

func TestEnvCheckReportsWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_EC_ONE=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_EC_ONE=1\nGOCFG_EC_TWO=2\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"env", "--check"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	if !strings.Contains(sb.String(), "would add GOCFG_EC_TWO") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".env"))
	if string(b) != "GOCFG_EC_ONE=1\n" {
		t.Fatalf("check wrote to file: %q", b)
	}
}

func TestEnvInSyncIsQuietSuccess(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_EQ_ONE=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_EQ_ONE=1\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"env"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), "in sync") {
		t.Fatalf("unexpected output:\n%s", sb.String())
	}
}
