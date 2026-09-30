package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCreatesAllFiles(t *testing.T) {
	dir := t.TempDir()
	res, err := Run(Options{Dir: dir, AppName: "test-app"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 4 {
		t.Fatalf("got %d files: %+v", len(res.Files), res.Files)
	}
	for _, f := range res.Files {
		if !f.Created || f.Skipped {
			t.Fatalf("unexpected result %+v", f)
		}
	}
	for _, rel := range []string{".env", ".env.example", "config/app.go", "config/config.go"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".env"))
	if string(data) != "APP_NAME=test-app\nAPP_ENV=dev\nAPP_DEBUG=true\nAPP_URL=http://localhost:9000\n" {
		t.Fatalf("unexpected .env:\n%s", data)
	}
	if fi, _ := os.Stat(filepath.Join(dir, ".env")); fi.Mode().Perm() != 0o600 {
		t.Fatalf(".env perm = %o, want 600", fi.Mode().Perm())
	}
}

func TestRunSkipsExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(Options{Dir: dir}); err != nil {
		t.Fatal(err)
	}
	marker := []byte("APP_NAME=keep-me\n")
	if err := os.WriteFile(filepath.Join(dir, ".env"), marker, 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := Run(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Files {
		if f.Path == ".env" && !f.Skipped {
			t.Fatalf(".env should be skipped: %+v", f)
		}
	}
	got, _ := os.ReadFile(filepath.Join(dir, ".env"))
	if string(got) != string(marker) {
		t.Fatalf(".env overwritten: %q", got)
	}
}

func TestRunForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(Options{Dir: dir}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("custom\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := Run(Options{Dir: dir, Force: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Files {
		if f.Skipped {
			t.Fatalf("nothing should be skipped with force: %+v", f)
		}
	}
	got, _ := os.ReadFile(filepath.Join(dir, ".env"))
	if string(got) == "custom\n" {
		t.Fatal(".env not overwritten")
	}
}

func TestRunDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	res, err := Run(Options{Dir: dir, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.DryRun {
		t.Fatal("expected DryRun result")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("dry-run wrote files: %v", entries)
	}
}
