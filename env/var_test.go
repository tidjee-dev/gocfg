package env

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestVarDefaults(t *testing.T) {
	if v, err := StringVar("GOCFG_VD_STR", "dflt").Resolve(); err != nil || v != "dflt" {
		t.Fatalf("got %q, %v", v, err)
	}
	if v, err := IntVar("GOCFG_VD_INT", 7).Resolve(); err != nil || v != 7 {
		t.Fatalf("got %d, %v", v, err)
	}
	if v, err := BoolVar("GOCFG_VD_BOOL", true).Resolve(); err != nil || !v {
		t.Fatalf("got %v, %v", v, err)
	}
	if v, err := DurationVar("GOCFG_VD_DUR", time.Second).Resolve(); err != nil || v != time.Second {
		t.Fatalf("got %v, %v", v, err)
	}
}

func TestVarPrecedence(t *testing.T) {
	t.Setenv("GOCFG_VP", "os")
	if v, err := StringVar("GOCFG_VP", "dflt").Resolve(); err != nil || v != "os" {
		t.Fatalf("OS must win: got %q, %v", v, err)
	}
}

func TestVarEmptyMeansUnset(t *testing.T) {
	t.Setenv("GOCFG_VE_INT", "")
	if v, err := IntVar("GOCFG_VE_INT", 42).Resolve(); err != nil || v != 42 {
		t.Fatalf("empty optional must yield default: got %d, %v", v, err)
	}
	t.Setenv("GOCFG_VE_STR", "")
	if v, err := StringVar("GOCFG_VE_STR", "dflt").Resolve(); err != nil || v != "dflt" {
		t.Fatalf("empty optional must yield default: got %q, %v", v, err)
	}
}

func TestVarRequired(t *testing.T) {
	if _, err := StringVar("GOCFG_VR_MISSING", "", Require()).Resolve(); err == nil {
		t.Fatal("expected RequiredError for missing")
	} else {
		var rerr *RequiredError
		if !errors.As(err, &rerr) {
			t.Fatalf("expected *RequiredError, got %v", err)
		}
	}
	t.Setenv("GOCFG_VR_EMPTY", "")
	if _, err := StringVar("GOCFG_VR_EMPTY", "", Require()).Resolve(); err == nil {
		t.Fatal("expected RequiredError for empty")
	}
	t.Setenv("GOCFG_VR_SET", "yes")
	if v, err := StringVar("GOCFG_VR_SET", "", Require()).Resolve(); err != nil || v != "yes" {
		t.Fatalf("got %q, %v", v, err)
	}
}

func TestVarParseErrorMatchesGetter(t *testing.T) {
	t.Setenv("GOCFG_VG", "hello")
	_, verr := IntVar("GOCFG_VG", 0).Resolve()
	_, gerr := Int("GOCFG_VG", 0)
	if verr == nil || gerr == nil {
		t.Fatal("both must fail")
	}
	if verr.Error() != gerr.Error() {
		t.Fatalf("var %q != getter %q", verr, gerr)
	}
}

func TestVarSecretOptRedacts(t *testing.T) {
	// Key matches no heuristic marker.
	t.Setenv("GOCFG_VS_PLAIN", "nope-not-a-bool")
	_, err := BoolVar("GOCFG_VS_PLAIN", false, Secret()).Resolve()
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "nope-not-a-bool") {
		t.Fatalf("secret leaked: %v", err)
	}
	var perr *Error
	if !errors.As(err, &perr) || !perr.Secret {
		t.Fatalf("expected secret *Error, got %v", err)
	}
	// Without the opt, the same key quotes the value.
	_, err = BoolVar("GOCFG_VS_PLAIN2", false).Resolve()
	t.Setenv("GOCFG_VS_PLAIN2", "also-bad")
	_, err = BoolVar("GOCFG_VS_PLAIN2", false).Resolve()
	if err == nil || !strings.Contains(err.Error(), `"also-bad"`) {
		t.Fatalf("expected quoted value, got %v", err)
	}
}

func TestVarHeuristicStillRedacts(t *testing.T) {
	t.Setenv("GOCFG_VH_PASSWORD", "s3cr3t-bad-int")
	_, err := IntVar("GOCFG_VH_PASSWORD", 0).Resolve()
	if err == nil || strings.Contains(err.Error(), "s3cr3t-bad-int") {
		t.Fatalf("heuristic must redact: %v", err)
	}
}

func TestVarCustomParser(t *testing.T) {
	upper := StringVar("GOCFG_VC", "", Require())
	upper.Parse = func(key, value string) (string, error) {
		if value == "boom" {
			return "", &Error{Key: key, Kind: "custom", Value: value}
		}
		return strings.ToUpper(value), nil
	}
	t.Setenv("GOCFG_VC", "hi")
	if v, err := upper.Resolve(); err != nil || v != "HI" {
		t.Fatalf("got %q, %v", v, err)
	}
	t.Setenv("GOCFG_VC", "boom")
	if _, err := upper.Resolve(); err == nil {
		t.Fatal("expected custom error")
	}
	// Secret var with leaking custom parser: message replaced.
	leaky := StringVar("GOCFG_VC2_PLAIN", "", Secret())
	leaky.Parse = func(key, value string) (string, error) {
		return "", errors.New("leaked-" + value)
	}
	t.Setenv("GOCFG_VC2_PLAIN", "s3cr3t")
	_, err := leaky.Resolve()
	if err == nil || strings.Contains(err.Error(), "s3cr3t") {
		t.Fatalf("must redact custom parser error: %v", err)
	}
}

func TestAnyConformance(t *testing.T) {
	defs := []Any{
		StringVar("GOCFG_VA_A", "dflt"),
		BoolVar("GOCFG_VA_B", true),
		IntVar("GOCFG_VA_C", 1),
		Int64Var("GOCFG_VA_D", 2),
		Float64Var("GOCFG_VA_E", 1.5),
		DurationVar("GOCFG_VA_F", time.Second, Secret(), Require()),
	}
	kinds := map[string]string{
		"GOCFG_VA_A": "string", "GOCFG_VA_B": "bool", "GOCFG_VA_C": "int",
		"GOCFG_VA_D": "int64", "GOCFG_VA_E": "float64", "GOCFG_VA_F": "duration",
	}
	for _, d := range defs {
		if d.AnyKind() != kinds[d.AnyKey()] {
			t.Fatalf("%s kind = %q", d.AnyKey(), d.AnyKind())
		}
		if d.IsSecret() != (d.AnyKey() == "GOCFG_VA_F") {
			t.Fatalf("%s secret = %v", d.AnyKey(), d.IsSecret())
		}
		if d.IsRequired() != (d.AnyKey() == "GOCFG_VA_F") {
			t.Fatalf("%s required = %v", d.AnyKey(), d.IsRequired())
		}
	}
	if got := StringVar("K", "dflt").AnyDefault(); got != "dflt" {
		t.Fatalf("default = %q", got)
	}
	if got := DurationVar("K", 90*time.Second).AnyDefault(); got != "1m30s" {
		t.Fatalf("default = %q", got)
	}
}
