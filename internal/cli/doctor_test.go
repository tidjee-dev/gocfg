package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorNarrative(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Duplicate key, OS shadowing, secret value in example, no .gitignore.
	writeNamedEnv(t, dir, ".env", "GOCFG_DR_A=file\nGOCFG_DR_A=dup\nGOCFG_DR_B=file\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_DR_A=\nGOCFG_DR_B=\nGOCFG_DR_PASSWORD=hunter2\n", 0o644)
	t.Setenv("GOCFG_DR_B", "os")
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"doctor"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doctor always exits 0, got %v", err)
	}
	out := sb.String()
	for _, want := range []string{
		"gocfg doctor",
		"toolchain: go",
		"✓ config/ present",
		".env parses (2 keys)",
		"no .gitignore",
		"duplicate key in .env: GOCFG_DR_A (2x, last wins)",
		"1 OS variable(s) shadow .env: GOCFG_DR_B",
		"1 secret(s) have values in .env.example: GOCFG_DR_PASSWORD",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "hunter2") {
		t.Fatalf("secret leaked:\n%s", out)
	}
}

func TestDoctorGitignoreCovered(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "A=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "A=1\n", 0o644)
	writeNamedEnv(t, dir, ".gitignore", ".env\n", 0o644)
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"doctor"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sb.String(), ".gitignore") {
		t.Fatalf("covered .env must stay quiet:\n%s", sb.String())
	}
}
