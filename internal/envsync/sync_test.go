package envsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tidjee-dev/gocfg/env"
)

func writeFile(t *testing.T, dir, name, content string, perm os.FileMode) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), perm); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestCreatesMissingEnv(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env.example", "APP_NAME=My App\nAPP_ENV=dev\n", 0o644)
	res, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created || !res.Changed || len(res.Added) != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}
	got := readFile(t, dir, ".env")
	if got != "APP_ENV=dev\nAPP_NAME=My App\n" {
		t.Fatalf("unexpected .env:\n%s", got)
	}
	if fi, _ := os.Stat(filepath.Join(dir, ".env")); fi.Mode().Perm() != 0o600 {
		t.Fatalf(".env perm = %o, want 600", fi.Mode().Perm())
	}
}

func TestAppendsMissingPreservingAll(t *testing.T) {
	dir := t.TempDir()
	before := "# my production file\nAPP_NAME=My Production App\n\nAPP_ENV=prod   \n"
	writeFile(t, dir, ".env", before, 0o600)
	writeFile(t, dir, ".env.example", "APP_NAME=x\nAPP_ENV=x\nAPP_NEW=hello\n", 0o644)
	res, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Created || len(res.Added) != 1 || res.Added[0] != "APP_NEW" {
		t.Fatalf("unexpected result: %+v", res)
	}
	got := readFile(t, dir, ".env")
	want := before + "APP_NEW=hello\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "A=1", 0o600)
	writeFile(t, dir, ".env.example", "A=1\nB=2\n", 0o644)
	if _, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, dir, ".env"); got != "A=1\nB=2\n" {
		t.Fatalf("got %q", got)
	}
}

func TestSecretsAppendedEmptyWithWarning(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "A=1\n", 0o600)
	writeFile(t, dir, ".env.example", "A=1\nDB_PASSWORD=hunter2\nAPI_KEY=sk-live-42\n", 0o644)
	res, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")})
	if err != nil {
		t.Fatal(err)
	}
	got := readFile(t, dir, ".env")
	if !strings.Contains(got, "DB_PASSWORD=\n") || !strings.Contains(got, "API_KEY=\n") {
		t.Fatalf("secrets must be appended empty:\n%s", got)
	}
	if strings.Contains(got, "hunter2") || strings.Contains(got, "sk-live-42") {
		t.Fatalf("secret value leaked into .env:\n%s", got)
	}
	for _, w := range res.Warnings {
		if strings.Contains(w, "hunter2") || strings.Contains(w, "sk-live-42") {
			t.Fatalf("secret value leaked into warning: %q", w)
		}
	}
	if len(res.Warnings) != 2 {
		t.Fatalf("expected 2 secret warnings, got %+v", res.Warnings)
	}
}

func TestExtrasWarnOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "A=1\nEXTRA=9\n", 0o600)
	writeFile(t, dir, ".env.example", "A=1\n", 0o644)
	res, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Changed {
		t.Fatalf("extras must not change the file: %+v", res)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "EXTRA") {
		t.Fatalf("expected extra warning, got %+v", res.Warnings)
	}
	if got := readFile(t, dir, ".env"); got != "A=1\nEXTRA=9\n" {
		t.Fatalf("file touched: %q", got)
	}
}

func TestCheckWritesNothing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "A=1\n", 0o600)
	writeFile(t, dir, ".env.example", "A=1\nB=2\n", 0o644)
	res, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example"), Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed || len(res.Added) != 1 {
		t.Fatalf("check must report change: %+v", res)
	}
	if got := readFile(t, dir, ".env"); got != "A=1\n" {
		t.Fatalf("check wrote to file: %q", got)
	}
}

func TestMissingExampleErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")}); err == nil {
		t.Fatal("expected error for missing example")
	}
}

func TestMalformedFilesError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "BROKEN LINE\n", 0o600)
	writeFile(t, dir, ".env.example", "A=1\n", 0o644)
	if _, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")}); err == nil {
		t.Fatal("expected error for malformed .env")
	}
	writeFile(t, dir, ".env", "A=1\n", 0o600)
	writeFile(t, dir, ".env.example", "BROKEN LINE\n", 0o644)
	if _, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")}); err == nil {
		t.Fatal("expected error for malformed example")
	}
}

func TestExistingModePreserved(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "A=1\n", 0o640)
	writeFile(t, dir, ".env.example", "A=1\nB=2\n", 0o644)
	if _, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")}); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(filepath.Join(dir, ".env")); fi.Mode().Perm() != 0o640 {
		t.Fatalf(".env perm changed to %o", fi.Mode().Perm())
	}
}

func TestFailedWriteLeavesOriginalIntact(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "A=1\n", 0o600)
	writeFile(t, dir, ".env.example", "A=1\nB=2\n", 0o644)
	// Read-only dir: temp file creation fails, original must survive
	// with no stray temp files left behind.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if _, err := Run(Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")}); err == nil {
		t.Fatal("expected error for unwritable dir")
	}
	if got := readFile(t, dir, ".env"); got != "A=1\n" {
		t.Fatalf("original modified: %q", got)
	}
	leftovers, err := filepath.Glob(filepath.Join(dir, ".gocfg-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(leftovers) != 0 {
		t.Fatalf("stray temp files: %v", leftovers)
	}
}

func TestRunDefsPreservesManifestOrder(t *testing.T) {
	dir := t.TempDir()
	defs := []env.Definition{
		{Key: "GOCFG_D_ZULU", Kind: "string", Default: "z"},
		{Key: "GOCFG_D_ALPHA", Kind: "int", Default: "1"},
		{Key: "GOCFG_D_MIKE", Kind: "string", Default: "m"},
	}
	res, err := RunDefs(Options{EnvFile: filepath.Join(dir, ".env")}, defs)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created || len(res.Added) != 3 {
		t.Fatalf("unexpected result: %+v", res)
	}
	got := readFile(t, dir, ".env")
	want := "GOCFG_D_ZULU=z\nGOCFG_D_ALPHA=1\nGOCFG_D_MIKE=m\n"
	if got != want {
		t.Fatalf("manifest order not preserved:\n%q\nwant:\n%q", got, want)
	}
}

func TestRunDefsSecretsAndRequiredEmpty(t *testing.T) {
	dir := t.TempDir()
	defs := []env.Definition{
		{Key: "GOCFG_DD_PLAIN", Kind: "string", Default: "v"},
		{Key: "GOCFG_DD_PASSWORD", Kind: "string", Default: "hunter2"},
		{Key: "GOCFG_DD_REQ", Kind: "int", Required: true},
	}
	res, err := RunDefs(Options{EnvFile: filepath.Join(dir, ".env")}, defs)
	if err != nil {
		t.Fatal(err)
	}
	got := readFile(t, dir, ".env")
	want := "GOCFG_DD_PLAIN=v\nGOCFG_DD_PASSWORD=\nGOCFG_DD_REQ=\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
	if strings.Contains(got, "hunter2") {
		t.Fatalf("secret leaked:\n%s", got)
	}
	found := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "hunter2") {
			t.Fatalf("secret leaked into warning: %q", w)
		}
		if strings.Contains(w, "GOCFG_DD_PASSWORD") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected secret warning, got %+v", res.Warnings)
	}
}
