package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tidjee-dev/gocfg/env"
)

func writeManifest(t *testing.T, dir string, defs []env.Any) string {
	t.Helper()
	raw, err := env.MarshalDefinitions(defs)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, ".gocfg.json")
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

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

func writeNamedEnv(t *testing.T, dir, name, content string, perm os.FileMode) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), perm); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAllSuccess(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_VA_ONE=1\nGOCFG_VA_TWO=2\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_VA_ONE=\nGOCFG_VA_TWO=\n", 0o644)
	out, err := runValidate(t, dir, "validate")
	if err != nil {
		t.Fatalf("err: %v\nout:\n%s", err, out)
	}
	for _, want := range []string{"✓ GOCFG_VA_ONE", "✓ GOCFG_VA_TWO", "Configuration is valid."} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestValidateAllMissing(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_VB_ONE=1\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_VB_ONE=\nGOCFG_VB_TWO=\n", 0o644)
	out, err := runValidate(t, dir, "validate")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v\nout:\n%s", err, out)
	}
	for _, want := range []string{"✓ GOCFG_VB_ONE", "✗ GOCFG_VB_TWO", "Configuration is invalid."} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestValidateAllFromOSWithoutEnvFile(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env.example", "GOCFG_VC_ONLY=\n", 0o644)
	t.Setenv("GOCFG_VC_ONLY", "from-os")
	out, err := runValidate(t, dir, "validate")
	if err != nil {
		t.Fatalf("err: %v\nout:\n%s", err, out)
	}
	if !strings.Contains(out, "✓ GOCFG_VC_ONLY") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestValidateAllMalformedEnv(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "BROKEN LINE\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_VD_ONE=\n", 0o644)
	_, err := runValidate(t, dir, "validate")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
}

func TestValidateAllNoExample(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_VE_ONE=1\n", 0o600)
	_, err := runValidate(t, dir, "validate")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v", err)
	}
}

func TestValidateAllExtraWarns(t *testing.T) {
	dir := t.TempDir()
	writeNamedEnv(t, dir, ".env", "GOCFG_VF_ONE=1\nGOCFG_VF_EXTRA=9\n", 0o600)
	writeNamedEnv(t, dir, ".env.example", "GOCFG_VF_ONE=\n", 0o644)
	out, err := runValidate(t, dir, "validate")
	if err != nil {
		t.Fatalf("extras must not fail: %v\nout:\n%s", err, out)
	}
	if !strings.Contains(out, "! GOCFG_VF_EXTRA (not in .env.example)") {
		t.Fatalf("missing extra warning in:\n%s", out)
	}
}

func TestValidateDefsTyped(t *testing.T) {
	dir := t.TempDir()
	m := writeManifest(t, dir, []env.Any{
		env.IntVar("GOCFG_VD_PORT", 9000),
		env.BoolVar("GOCFG_VD_DEBUG", false),
		env.StringVar("GOCFG_VD_REQ", "", env.Require()),
	})
	p := filepath.Join(dir, ".env")
	writeNamedEnv(t, dir, ".env", "GOCFG_VD_PORT=not-a-number\nGOCFG_VD_DEBUG=true\n", 0o600)
	out, err := runValidate(t, dir, "validate", "--env-file", p, "--defs", m)
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Fatalf("expected ExitError{1}, got %v\nout:\n%s", err, out)
	}
	for _, want := range []string{"✗ GOCFG_VD_PORT", "✓ GOCFG_VD_DEBUG", "✗ GOCFG_VD_REQ"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestValidateDefsUnknownKind(t *testing.T) {
	dir := t.TempDir()
	m := filepath.Join(dir, ".gocfg.json")
	os.WriteFile(m, []byte(`{"version":1,"vars":[{"key":"GOCFG_VK","kind":"future"}]}`), 0o644)
	p := filepath.Join(dir, ".env")
	writeNamedEnv(t, dir, ".env", "GOCFG_VK=x\n", 0o600)
	out, err := runValidate(t, dir, "validate", "--env-file", p, "--defs", m)
	if err == nil || !strings.Contains(out, "unsupported kind") {
		t.Fatalf("expected unsupported-kind error, got %v\nout:\n%s", err, out)
	}
}
