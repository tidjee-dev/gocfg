package dotenv

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestGodotenvFixtures parses snapshots of joho/godotenv's fixtures
// (see testdata/ATTRIBUTION.md) and asserts our behavior.
//
// Intentional deltas from godotenv, covered below:
//   - keys are ASCII [A-Za-z_][A-Za-z0-9_.-]* (no leading digit, no ':' separator)
//   - ${VAR} expansion prefers OS env over earlier file keys (OS > .env)
//   - \t maps to tab (godotenv yields "t")
func TestGodotenvFixtures(t *testing.T) {
	t.Setenv("GLOBAL_OPTION", "global")

	tests := []struct {
		file string
		want map[string]string
	}{
		{"godotenv-comments.env", map[string]string{
			"qux": "thud", "thud": "fred#qux", "fred": "qux#baz",
			"foo": "bar", "bar": "foo#baz", "baz": "foo",
		}},
		{"godotenv-equals.env", map[string]string{
			"OPTION_A": "postgres://localhost:5432/database?sslmode=disable",
		}},
		{"godotenv-exported.env", map[string]string{
			"OPTION_A": "2", "OPTION_B": `\n`,
		}},
		{"godotenv-hyphen.env", map[string]string{
			"OPTION_A": "abc", "OPTION-B": "def",
		}},
		{"godotenv-plain.env", map[string]string{
			"OPTION_A": "1", "OPTION_B": "2", "OPTION_C": "3",
			"OPTION_D": "4", "OPTION_E": "5", "OPTION_F": "",
			"OPTION_G": "", "OPTION_H": "1 2",
		}},
		{"godotenv-quoted.env", map[string]string{
			"OPTION_A": "1", "OPTION_B": "2", "OPTION_C": "",
			"OPTION_D": `\n`, "OPTION_E": "1", "OPTION_F": "2",
			"OPTION_G": "", "OPTION_H": "\n",
			"OPTION_I": "echo 'asd'",
			"OPTION_J": "line 1\nline 2",
			"OPTION_K": `line one` + "\n" + `this is \'quoted\'` + "\n" + `one more line`,
			"OPTION_L": "line 1\nline 2",
			"OPTION_M": "line one\nthis is \"quoted\"\none more line",
		}},
		{"godotenv-substitutions.env", map[string]string{
			"OPTION_A": "1", "OPTION_B": "1", "OPTION_C": "1",
			"OPTION_D": "11", "OPTION_E": "", "OPTION_F": "global",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			got, err := Parse(raw)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGodotenvInvalidFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "godotenv-invalid1.env"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(raw)
	var perr *ParseError
	if !errors.As(err, &perr) || perr.Line != 1 {
		t.Fatalf("expected *ParseError line 1, got %v", err)
	}
}

// TestExpansionPrefersOS documents the delta: godotenv prefers earlier
// file keys, we prefer OS env (consistent with OS > .env precedence).
func TestExpansionPrefersOS(t *testing.T) {
	t.Setenv("OPTION_A", "os")
	got, err := Parse([]byte("OPTION_A=file\nOPTION_B=${OPTION_A}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got["OPTION_B"] != "os" {
		t.Fatalf("got %q, want os-first expansion", got["OPTION_B"])
	}
}
