package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEnv(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadNeverOverridesOS(t *testing.T) {
	t.Setenv("APP_PORT", "7000")
	p := writeEnv(t, "APP_PORT=8000\nAPP_NEW=hello\n")
	if err := Load(p); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("APP_PORT") != "7000" {
		t.Fatalf("OS value overridden: %q", os.Getenv("APP_PORT"))
	}
	if os.Getenv("APP_NEW") != "hello" {
		t.Fatalf("new key not set: %q", os.Getenv("APP_NEW"))
	}
}

func TestLoadDefaultMissingIsNoop(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := Load(); err != nil {
		t.Fatalf("missing default .env should be nil, got %v", err)
	}
}

func TestLoadExplicitMissingErrors(t *testing.T) {
	p := filepath.Join(t.TempDir(), "does-not-exist.env")
	if err := Load(p); err == nil {
		t.Fatal("expected error for missing explicit path")
	}
}

func TestLoadMalformedPropagatesLineError(t *testing.T) {
	p := writeEnv(t, "GOCFG_S1_OK=1\nBROKEN LINE\n")
	err := Load(p)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if perr, ok := err.(*ParseError); !ok || perr.Line != 2 {
		t.Fatalf("expected *ParseError line 2, got %T (%v)", err, err)
	}
	// Nothing from the file may leak into the environment on failure.
	if _, ok := os.LookupEnv("GOCFG_S1_OK"); ok {
		// Parse-all-then-apply guarantees atomicity per file.
		t.Fatal("partial application: GOCFG_S1_OK must not be set after failed load")
	}
}

func TestLoadMultiPath(t *testing.T) {
	a := writeEnv(t, "A=1\nB=from-a\n")
	b := writeEnv(t, "B=from-b\nC=3\n")
	t.Setenv("C", "os")
	if err := Load(a, b); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("A") != "1" || os.Getenv("B") != "from-a" || os.Getenv("C") != "os" {
		t.Fatalf("A=%q B=%q C=%q", os.Getenv("A"), os.Getenv("B"), os.Getenv("C"))
	}
}
