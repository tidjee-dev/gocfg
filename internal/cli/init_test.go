package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
}

func TestInitCreatesProject(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"init", "--name", "demo"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".env", ".env.example", "config/app.go", "config/config.go", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
	}
	if !strings.Contains(sb.String(), "created .env") {
		t.Fatalf("output missing creations:\n%s", sb.String())
	}
}

func TestInitGitignoreOptOut(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"init", "--gitignore=false"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Fatal("--gitignore=false must not touch .gitignore")
	}
}

func TestInitSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	run := func(args ...string) string {
		root := newRootCmd()
		var sb strings.Builder
		root.SetOut(&sb)
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		return sb.String()
	}
	run("init")
	out := run("init")
	if !strings.Contains(out, "already exists — skipped") {
		t.Fatalf("expected skip message:\n%s", out)
	}
}
