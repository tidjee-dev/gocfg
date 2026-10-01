package cli

import (
	"errors"
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
	for _, rel := range []string{".env", ".env.example", "internal/config/app.go", "internal/config/config.go", ".gitignore"} {
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

func TestInitWithDefinitions(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs([]string{"init", "--with-definitions"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"internal/config/vars.go", "tools/gocfg-gen/main.go"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
	}
	out := sb.String()
	if !strings.Contains(out, "created internal/config/vars.go") || !strings.Contains(out, "gocfg-gen > .gocfg.json") {
		t.Fatalf("unexpected output:\n%s", out)
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

func TestInitCustomConfigDirEndToEnd(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	run := func(args ...string) (string, error) {
		root := newRootCmd()
		var sb strings.Builder
		root.SetOut(&sb)
		root.SetArgs(args)
		err := root.Execute()
		return sb.String(), err
	}
	if out, err := run("init", "--config-dir", "custom"); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	} else if !strings.Contains(out, "created custom/app.go") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if out, err := run("check", "--config-dir", "custom"); err != nil {
		t.Fatalf("check: %v\nout:\n%s", err, out)
	}
}

func TestInitRejectsEscapingConfigDir(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	root := newRootCmd()
	root.SetArgs([]string{"init", "--config-dir", "../evil"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected usage ExitError{2}, got %v", err)
	}
}
