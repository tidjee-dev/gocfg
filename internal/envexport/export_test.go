package envexport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setup(t *testing.T, env, example string) Options {
	t.Helper()
	dir := t.TempDir()
	if env != "" {
		if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte(example), 0o644); err != nil {
		t.Fatal(err)
	}
	return Options{EnvFile: filepath.Join(dir, ".env"), ExampleFile: filepath.Join(dir, ".env.example")}
}

func TestCollectOSWins(t *testing.T) {
	t.Setenv("GOCFG_X_A", "os")
	entries, err := Collect(setup(t, "GOCFG_X_A=file\nGOCFG_X_B=file\n", "GOCFG_X_A=\nGOCFG_X_B=\nGOCFG_X_C=\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]Entry{
		"GOCFG_X_A": {Key: "GOCFG_X_A", Value: "os"},
		"GOCFG_X_B": {Key: "GOCFG_X_B", Value: "file"},
		"GOCFG_X_C": {Key: "GOCFG_X_C", Missing: true},
	}
	if len(entries) != 3 {
		t.Fatalf("got %+v", entries)
	}
	for _, e := range entries {
		if w := want[e.Key]; e.Value != w.Value || e.Missing != w.Missing {
			t.Fatalf("%s: got %+v, want %+v", e.Key, e, w)
		}
	}
}

func TestCollectIgnoresExtras(t *testing.T) {
	entries, err := Collect(setup(t, "EXTRA=9\n", "A=1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Key != "A" {
		t.Fatalf("extras must not be exported: %+v", entries)
	}
}

func TestFormatShellEscapes(t *testing.T) {
	got := FormatShell([]Entry{{Key: "A", Value: "o'clock"}, {Key: "B", Value: "plain"}})
	want := "export A='o'\\''clock'\nexport B='plain'\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	// Newlines stay literal inside single quotes (valid POSIX).
	if got := FormatShell([]Entry{{Key: "A", Value: "x\ny"}}); got != "export A='x\ny'\n" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatDotenvQuotes(t *testing.T) {
	got := FormatDotenv([]Entry{
		{Key: "PLAIN", Value: "abc"},
		{Key: "SPACED", Value: " a "},
		{Key: "MULTI", Value: "x\ny"},
		{Key: "QUOTE", Value: `say "hi"`},
	})
	for _, want := range []string{
		"PLAIN=abc\n",
		"SPACED=\" a \"\n",
		"MULTI=\"x\\ny\"\n",
		"QUOTE=\"say \\\"hi\\\"\"\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatJSON(t *testing.T) {
	got, err := FormatJSON([]Entry{{Key: "A", Value: "1"}, {Key: "B", Value: ""}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"A": "1"`) || !strings.Contains(got, `"B": ""`) {
		t.Fatalf("unexpected json:\n%s", got)
	}
}
