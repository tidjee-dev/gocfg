package envdiff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tidjee-dev/gocfg/env"
)

func setup(t *testing.T, envFile, example string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if envFile != "" {
		if err := os.WriteFile(envPath, []byte(envFile), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	examplePath := filepath.Join(dir, ".env.example")
	if err := os.WriteFile(examplePath, []byte(example), 0o644); err != nil {
		t.Fatal(err)
	}
	return envPath, examplePath
}

func byKey(rows []Row) map[string]Row {
	m := make(map[string]Row, len(rows))
	for _, r := range rows {
		m[r.Key] = r
	}
	return m
}

func TestTwoWayClean(t *testing.T) {
	envPath, examplePath := setup(t, "A=1\n", "A=1\n")
	rows, err := Compare(envPath, examplePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Status != Ok {
		t.Fatalf("got %+v", rows)
	}
}

func TestTwoWayMissingAndExtra(t *testing.T) {
	envPath, examplePath := setup(t, "A=1\nEXTRA=9\n", "A=1\nB=2\n")
	rows, err := Compare(envPath, examplePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := byKey(rows)
	if m["B"].Status != Fail || m["EXTRA"].Status != Warn || m["A"].Status != Ok {
		t.Fatalf("got %+v", rows)
	}
}

func TestThreeWayDrift(t *testing.T) {
	envPath, examplePath := setup(t, "A=1\n", "A=1\nB=old\n")
	defs := []env.Definition{{Key: "A", Kind: "string", Default: "1"}, {Key: "B", Kind: "string", Default: "new"}}
	rows, err := Compare(envPath, examplePath, defs)
	if err != nil {
		t.Fatal(err)
	}
	m := byKey(rows)
	if m["B"].Status != Fail {
		t.Fatalf("expected drift failure, got %+v", rows)
	}
}

func TestManifestOnlyKeyFails(t *testing.T) {
	envPath, examplePath := setup(t, "A=1\n", "A=1\n")
	defs := []env.Definition{{Key: "A", Kind: "string", Default: "1"}, {Key: "NEW", Kind: "string", Default: "n"}}
	rows, err := Compare(envPath, examplePath, defs)
	if err != nil {
		t.Fatal(err)
	}
	if byKey(rows)["NEW"].Status != Fail {
		t.Fatalf("got %+v", rows)
	}
}

func TestSecretsNeverLeak(t *testing.T) {
	envPath, examplePath := setup(t, "DB_PASSWORD=hunter2\n", "DB_PASSWORD=hunter2\n")
	rows, err := Compare(envPath, examplePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if strings.Contains(r.Detail, "hunter2") {
			t.Fatalf("secret leaked in row %+v", r)
		}
		if !r.Secret {
			t.Fatalf("expected secret flag on %+v", r)
		}
	}
}
