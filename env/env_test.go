package env

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStringFallbackAndValue(t *testing.T) {
	if v, _ := String("GOCFG_MISSING_STRING", "fb"); v != "fb" {
		t.Fatalf("fallback: got %q", v)
	}
	t.Setenv("GOCFG_STR", "hello")
	if v, err := String("GOCFG_STR", "fb"); err != nil || v != "hello" {
		t.Fatalf("value: got %q, %v", v, err)
	}
}

func TestBoolSets(t *testing.T) {
	trues := []string{"1", "true", "TRUE", " True ", "yes", "YES", "y", "Y", "on", "ON"}
	falses := []string{"0", "false", "FALSE", " False ", "no", "NO", "n", "N", "off", "OFF"}
	for _, s := range trues {
		t.Setenv("GOCFG_BOOL", s)
		v, err := Bool("GOCFG_BOOL", false)
		if err != nil || !v {
			t.Fatalf("%q: got %v, %v", s, v, err)
		}
	}
	for _, s := range falses {
		t.Setenv("GOCFG_BOOL", s)
		v, err := Bool("GOCFG_BOOL", true)
		if err != nil || v {
			t.Fatalf("%q: got %v, %v", s, v, err)
		}
	}
	if v, _ := Bool("GOCFG_MISSING_BOOL", true); !v {
		t.Fatal("fallback true expected")
	}
}

func TestBoolInvalid(t *testing.T) {
	t.Setenv("GOCFG_BOOL_BAD", "maybe")
	_, err := Bool("GOCFG_BOOL_BAD", false)
	var perr *Error
	if !errors.As(err, &perr) || perr.Key != "GOCFG_BOOL_BAD" {
		t.Fatalf("expected *Error, got %v", err)
	}
}

func TestInt(t *testing.T) {
	if v, _ := Int("GOCFG_MISSING_INT", 9000); v != 9000 {
		t.Fatalf("fallback: got %d", v)
	}
	t.Setenv("GOCFG_INT", "7000")
	if v, err := Int("GOCFG_INT", 9000); err != nil || v != 7000 {
		t.Fatalf("got %d, %v", v, err)
	}
	t.Setenv("GOCFG_INT", "hello")
	_, err := Int("GOCFG_INT", 9000)
	var perr *Error
	if !errors.As(err, &perr) {
		t.Fatalf("expected *Error, got %v", err)
	}
	if !strings.Contains(err.Error(), `"hello"`) {
		t.Fatalf("error must quote value: %v", err)
	}
}

func TestInt64(t *testing.T) {
	t.Setenv("GOCFG_I64", "123456789012")
	if v, err := Int64("GOCFG_I64", 0); err != nil || v != 123456789012 {
		t.Fatalf("got %d, %v", v, err)
	}
	t.Setenv("GOCFG_I64", "abc")
	if _, err := Int64("GOCFG_I64", 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestFloat64(t *testing.T) {
	t.Setenv("GOCFG_F", "1.5")
	if v, err := Float64("GOCFG_F", 0); err != nil || v != 1.5 {
		t.Fatalf("got %v, %v", v, err)
	}
	t.Setenv("GOCFG_F", "abc")
	if _, err := Float64("GOCFG_F", 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestDuration(t *testing.T) {
	t.Setenv("GOCFG_DUR", "1m30s")
	if v, err := Duration("GOCFG_DUR", 0); err != nil || v != 90*time.Second {
		t.Fatalf("got %v, %v", v, err)
	}
	t.Setenv("GOCFG_DUR", "not-a-duration")
	if _, err := Duration("GOCFG_DUR", 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestSecretRedaction(t *testing.T) {
	for _, key := range []string{"DB_PASSWORD", "APP_SECRET", "API_KEY", "AUTH_TOKEN", "APP_KEY"} {
		t.Setenv(key, "super-secret-value")
		_, err := Int(key, 0) // force a parse error on a secret key
		if err == nil {
			t.Fatalf("%s: expected error", key)
		}
		if strings.Contains(err.Error(), "super-secret-value") {
			t.Fatalf("%s: error leaks secret: %v", key, err)
		}
		var perr *Error
		if !errors.As(err, &perr) || !perr.Secret {
			t.Fatalf("%s: expected secret *Error, got %v", key, err)
		}
	}
	if !IsSecretKey("db_password") || IsSecretKey("APP_PORT") {
		t.Fatal("heuristic wrong")
	}
}

func TestRequired(t *testing.T) {
	t.Setenv("GOCFG_REQ", "present")
	if v, err := Required("GOCFG_REQ"); err != nil || v != "present" {
		t.Fatalf("got %q, %v", v, err)
	}
	if _, err := Required("GOCFG_MISSING_REQ"); err == nil {
		t.Fatal("expected error for missing")
	} else {
		var rerr *RequiredError
		if !errors.As(err, &rerr) {
			t.Fatalf("expected *RequiredError, got %v", err)
		}
	}
	t.Setenv("GOCFG_EMPTY_REQ", "")
	if _, err := Required("GOCFG_EMPTY_REQ"); err == nil {
		t.Fatal("expected error for empty")
	}
}
