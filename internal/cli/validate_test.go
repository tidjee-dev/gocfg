package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func runValidate(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	chdir(t, dir)
	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetArgs(args)
	err := root.Execute()
	return sb.String(), err
}

func TestValidateSuccess(t *testing.T) {
	dir := t.TempDir()
	p := writeEnv(t, "GOCFG_VOK_PORT=9000\nGOCFG_VOK_DEBUG=true\nGOCFG_VOK_SECRET=topsecret\n")
	_ = p
	// place .env where the command looks: use --env-file explicitly
	out, err := runValidate(t, dir,
		"validate", "--env-file", p,
		"--int", "GOCFG_VOK_PORT", "--bool", "GOCFG_VOK_DEBUG", "--required", "GOCFG_VOK_SECRET")
	if err != nil {
		t.Fatalf("err: %v\nout:\n%s", err, out)
	}
	for _, want := range []string{"✓ GOCFG_VOK_PORT", "✓ GOCFG_VOK_DEBUG", "✓ GOCFG_VOK_SECRET", "Configuration is valid."} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestValidateFailureExitCode(t *testing.T) {
	dir := t.TempDir()
	p := writeEnv(t, "GOCFG_VBAD_PORT=not-a-number\n")
	out, err := runValidate(t, dir, "validate", "--env-file", p, "--int", "GOCFG_VBAD_PORT", "--required", "GOCFG_V_MISSING")
	if err == nil {
		t.Fatal("expected error")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
	for _, want := range []string{"✗ GOCFG_VBAD_PORT", "✗ GOCFG_V_MISSING", "Configuration is invalid."} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	// invalid value must be quoted in the message
	if !strings.Contains(out, `"not-a-number"`) {
		t.Fatalf("expected quoted value in:\n%s", out)
	}
}

func TestValidateRedactsSecrets(t *testing.T) {
	dir := t.TempDir()
	p := writeEnv(t, "GOCFG_VREDACT_DB_PASSWORD=not-a-number\n")
	out, err := runValidate(t, dir, "validate", "--env-file", p, "--int", "GOCFG_VREDACT_DB_PASSWORD")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(out, "not-a-number") {
		t.Fatalf("secret leaked in output:\n%s", out)
	}
}

func TestValidateOSBeatsFile(t *testing.T) {
	dir := t.TempDir()
	p := writeEnv(t, "GOCFG_VOS_PORT=8000\n")
	t.Setenv("GOCFG_VOS_PORT", "7000")
	out, err := runValidate(t, dir, "validate", "--env-file", p, "--int", "GOCFG_VOS_PORT")
	if err != nil {
		t.Fatalf("err: %v\nout:\n%s", err, out)
	}
	if !strings.Contains(out, "✓ GOCFG_VOS_PORT") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}
