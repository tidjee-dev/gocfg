package env

import (
	"errors"
	"net/netip"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func mustURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestURL(t *testing.T) {
	for _, s := range []string{"https://example.com", "postgres://localhost/db", "redis://:x@h:1/0"} {
		t.Setenv("GOCFG_URL", s)
		u, err := URL("GOCFG_URL", "")
		if err != nil || u.String() != s {
			t.Fatalf("%s: got %v, %v", s, u, err)
		}
	}
	for _, s := range []string{"no-scheme", "", "://bad"} {
		t.Setenv("GOCFG_URL", s)
		if _, err := URL("GOCFG_URL", "https://fallback"); err == nil {
			t.Fatalf("%q: expected error", s)
		}
	}
	// Unset with unparseable fallback fails fast (programmer error).
	if _, err := URL("GOCFG_URL_MISSING", "no-scheme"); err == nil {
		t.Fatal("expected error for invalid fallback")
	}
}

func TestIP(t *testing.T) {
	for _, s := range []string{"127.0.0.1", "::1", "10.0.0.1", "fe80::1%eth0"} {
		t.Setenv("GOCFG_IP", s)
		a, err := IP("GOCFG_IP", netip.Addr{})
		if err != nil || a.String() != s {
			t.Fatalf("%s: got %v, %v", s, a, err)
		}
	}
	t.Setenv("GOCFG_IP", "not-an-ip")
	if _, err := IP("GOCFG_IP", netip.Addr{}); err == nil {
		t.Fatal("expected error")
	}
	// Unset returns the fallback untouched (even zero).
	if a, err := IP("GOCFG_IP_MISSING", netip.Addr{}); err != nil || a.IsValid() {
		t.Fatalf("got %v, %v", a, err)
	}
}

func TestSplitList(t *testing.T) {
	tests := map[string][]string{
		`a,b,c`:         {"a", "b", "c"},
		` a , b `:       {"a", "b"},
		`"a,b", c`:      {"a,b", "c"},
		`a,,b`:          {"a", "b"},
		`,`:             nil,
		``:              nil,
		`'a,b',c`:       {"a,b", "c"},
		`"unterminated`: {"unterminated"},
	}
	for in, want := range tests {
		if got := splitList(in); !reflect.DeepEqual(got, want) {
			t.Fatalf("%q: got %v, want %v", in, got, want)
		}
	}
}

func TestStringSlice(t *testing.T) {
	t.Setenv("GOCFG_SS", "https://a.com, https://b.com")
	got, err := StringSlice("GOCFG_SS", nil)
	if err != nil || !reflect.DeepEqual(got, []string{"https://a.com", "https://b.com"}) {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestBoolSlice(t *testing.T) {
	t.Setenv("GOCFG_BS", "true, no, 1")
	got, err := BoolSlice("GOCFG_BS", nil)
	if err != nil || !reflect.DeepEqual(got, []bool{true, false, true}) {
		t.Fatalf("got %v, %v", got, err)
	}
	t.Setenv("GOCFG_BS", "true, maybe")
	if _, err := BoolSlice("GOCFG_BS", nil); err == nil {
		t.Fatal("expected error")
	} else {
		var perr *Error
		if !errors.As(err, &perr) || perr.Key != "GOCFG_BS" {
			t.Fatalf("expected keyed *Error, got %v", err)
		}
	}
}

func TestNewVarKinds(t *testing.T) {
	u := URLVar("GOCFG_NV_URL", "https://example.com")
	if u.AnyKind() != "url" || u.AnyDefault() != "https://example.com" {
		t.Fatalf("got %q %q", u.AnyKind(), u.AnyDefault())
	}
	ip := IPVar("GOCFG_NV_IP", netip.MustParseAddr("127.0.0.1"))
	if ip.AnyKind() != "ip" || ip.AnyDefault() != "127.0.0.1" {
		t.Fatalf("got %q %q", ip.AnyKind(), ip.AnyDefault())
	}
	ss := StringSliceVar("GOCFG_NV_SS", []string{"a", "b"})
	if ss.AnyKind() != "stringslice" || ss.AnyDefault() != "a,b" {
		t.Fatalf("got %q %q", ss.AnyKind(), ss.AnyDefault())
	}
	bs := BoolSliceVar("GOCFG_NV_BS", []bool{true, false})
	if bs.AnyKind() != "boolslice" || bs.AnyDefault() != "true,false" {
		t.Fatalf("got %q %q", bs.AnyKind(), bs.AnyDefault())
	}
	// Invalid URL default renders empty instead of panicking.
	bad := URLVar("GOCFG_NV_BAD", "no-scheme")
	if bad.AnyDefault() != "" {
		t.Fatalf("got %q", bad.AnyDefault())
	}
	t.Setenv("GOCFG_NV_URL", "https://set.example")
	if got, err := u.Resolve(); err != nil || got.String() != "https://set.example" {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestVarSecretRedactionNewKinds(t *testing.T) {
	t.Setenv("GOCFG_NR_TOKEN", "not a url")
	_, err := URLVar("GOCFG_NR_TOKEN", "", Secret()).Resolve()
	if err == nil || strings.Contains(err.Error(), "not a url") {
		t.Fatalf("must redact: %v", err)
	}
}
