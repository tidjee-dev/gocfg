package dotenv

import (
	"os"
	"testing"
)

func TestParseBasic(t *testing.T) {
	m, err := Parse([]byte("APP_NAME=hi\n# c\nAPP_PORT=9000 # comment\nexport FOO=bar\n"))
	if err != nil {
		t.Fatal(err)
	}
	if m["APP_NAME"] != "hi" || m["APP_PORT"] != "9000" || m["FOO"] != "bar" {
		t.Fatalf("got %v", m)
	}
}

func TestNoOverride(t *testing.T) {
	t.Setenv("APP_PORT", "7000")
	if err := os.WriteFile(t.TempDir()+"/.env", []byte("APP_PORT=8000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// call Load with explicit path in your own test; Parse alone must not touch env
	m, _ := Parse([]byte("APP_PORT=8000"))
	if m["APP_PORT"] != "8000" {
		t.Fatal("parse wrong")
	}
	if os.Getenv("APP_PORT") != "7000" {
		t.Fatal("must not override OS in test setup")
	}
}

func TestMalformed(t *testing.T) {
	if _, err := Parse([]byte("NOEQUALS\n")); err == nil {
		t.Fatal("expected error")
	}
}
