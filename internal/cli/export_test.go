package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportShell(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_EP_A=file-a\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_EP_A=\nGOCFG_EP_B=\n", 0o644)
	t.Setenv("GOCFG_EP_A", "os-a")
	chdir(t, dir)

	root := newRootCmd()
	var out, errOut strings.Builder
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"export"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "export GOCFG_EP_A='os-a'\nexport GOCFG_EP_B=''\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
	if !strings.Contains(errOut.String(), "GOCFG_EP_B is not set") {
		t.Fatalf("missing stderr warning: %q", errOut.String())
	}
}

func TestExportJSON(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_EJ_A=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_EJ_A=\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var out strings.Builder
	root.SetOut(&out)
	root.SetArgs([]string{"export", "--format", "json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"GOCFG_EJ_A": "1"`) {
		t.Fatalf("unexpected stdout: %q", out.String())
	}
}

func TestExportBadFormat(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env.example", "A=1\n", 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("A=1\n"), 0o600)
	chdir(t, dir)

	root := newRootCmd()
	root.SetArgs([]string{"export", "--format", "yaml"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected usage ExitError{2}, got %v", err)
	}
}
