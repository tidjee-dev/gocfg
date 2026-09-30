package gocfg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tidjee-dev/gocfg/env"
)

func TestLoadEnvPrecedence(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	content := "GOCFG_PRECEDENCE_A=file-a\nGOCFG_PRECEDENCE_B=file-b\n"
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOCFG_PRECEDENCE_A", "os-a")

	if err := LoadEnv(p); err != nil {
		t.Fatal(err)
	}
	a, err := env.String("GOCFG_PRECEDENCE_A", "default")
	if err != nil || a != "os-a" {
		t.Fatalf("OS must win: got %q, %v", a, err)
	}
	b, err := env.String("GOCFG_PRECEDENCE_B", "default")
	if err != nil || b != "file-b" {
		t.Fatalf(".env must beat default: got %q, %v", b, err)
	}
	c, err := env.String("GOCFG_PRECEDENCE_C_MISSING", "default")
	if err != nil || c != "default" {
		t.Fatalf("fallback: got %q, %v", c, err)
	}
}
