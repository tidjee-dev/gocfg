package env

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMarshalRoundTrip(t *testing.T) {
	defs := []Any{
		StringVar("APP_NAME", "My App"),
		IntVar("APP_PORT", 9000),
		BoolVar("APP_DEBUG", true),
		DurationVar("APP_TIMEOUT", 30*time.Second, Secret(), Require()),
	}
	raw, err := MarshalDefinitions(defs)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalDefinitions(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []Definition{
		{Key: "APP_NAME", Kind: "string", Default: "My App"},
		{Key: "APP_PORT", Kind: "int", Default: "9000"},
		{Key: "APP_DEBUG", Kind: "bool", Default: "true"},
		{Key: "APP_TIMEOUT", Kind: "duration", Secret: true, Required: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestMarshalEmpty(t *testing.T) {
	raw, err := MarshalDefinitions(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"version": 1`) {
		t.Fatalf("envelope must carry version: %s", raw)
	}
	got, err := UnmarshalDefinitions(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %+v, want empty", got)
	}
}

func TestUnmarshalRejects(t *testing.T) {
	for name, doc := range map[string]string{
		"garbage":     `not json`,
		"wrong type":  `{"version": "1"}`,
		"bad version": `{"version": 99, "vars": []}`,
		"empty key":   `{"version": 1, "vars": [{"key": "", "kind": "int"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := UnmarshalDefinitions([]byte(doc)); err == nil {
				t.Fatalf("expected error for %s", doc)
			}
		})
	}
}

func TestUnmarshalAllowsUnknownKind(t *testing.T) {
	// Kind support is checked by consumers, not the envelope.
	got, err := UnmarshalDefinitions([]byte(`{"version": 1, "vars": [{"key": "X", "kind": "future"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "future" {
		t.Fatalf("got %+v", got)
	}
}
