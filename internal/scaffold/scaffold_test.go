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

func TestGitignoreCreatedWhenMissing(t *testing.T) {
	dir := t.TempDir()
	res, err := Run(Options{Dir: dir, Gitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	var found *FileResult
	for i, f := range res.Files {
		if f.Path == ".gitignore" {
			found = &res.Files[i]
		}
	}
	if found == nil || !found.Created {
		t.Fatalf("expected created .gitignore, got %+v", res.Files)
	}
	got, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(got) != ".env\n" {
		t.Fatalf("unexpected .gitignore: %q", got)
	}
}

func TestGitignoreAppendedPreservingContent(t *testing.T) {
	dir := t.TempDir()
	before := "# binaries\n/my-app\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(Options{Dir: dir, Gitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	updated := false
	for _, f := range res.Files {
		if f.Path == ".gitignore" && f.Updated {
			updated = true
		}
	}
	if !updated {
		t.Fatalf("expected updated .gitignore, got %+v", res.Files)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, ".gitignore")); string(got) != before+".env\n" {
		t.Fatalf("unexpected .gitignore: %q", got)
	}
}

func TestGitignoreSkippedWhenCovered(t *testing.T) {
	for _, content := range []string{".env\n", "bin/\n*.env\n", "# comment\n/.env\n"} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		res, err := Run(Options{Dir: dir, Gitignore: true})
		if err != nil {
			t.Fatal(err)
		}
		skipped := false
		for _, f := range res.Files {
			if f.Path == ".gitignore" && f.Skipped {
				skipped = true
			}
		}
		if !skipped {
			t.Fatalf("%q: expected skipped .gitignore, got %+v", content, res.Files)
		}
		if got, _ := os.ReadFile(filepath.Join(dir, ".gitignore")); string(got) != content {
			t.Fatalf("covered .gitignore modified: %q", got)
		}
	}
}

func TestGitignoreNoDuplicateOnRerun(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		if _, err := Run(Options{Dir: dir, Gitignore: true}); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(got) != ".env\n" {
		t.Fatalf("duplicate entry: %q", got)
	}
}

func TestGitignoreOptOut(t *testing.T) {
	dir := t.TempDir()
	res, err := Run(Options{Dir: dir, Gitignore: false})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Files {
		if f.Path == ".gitignore" {
			t.Fatalf("opt-out must skip .gitignore, got %+v", res.Files)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Fatal(".gitignore should not exist")
	}
}

func TestGitignoreDryRun(t *testing.T) {
	dir := t.TempDir()
	res, err := Run(Options{Dir: dir, Gitignore: true, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write .gitignore")
	}
	for _, f := range res.Files {
		if f.Path == ".gitignore" && f.Created {
			return
		}
	}
	t.Fatalf("dry-run must report .gitignore creation: %+v", res.Files)
}
