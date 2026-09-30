package dotenv

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestParseTable(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		want  map[string]string
		setup map[string]string // OS env for expansion tests
	}{
		{
			name: "basic",
			src:  "APP_NAME=hi\nAPP_PORT=9000\n",
			want: map[string]string{"APP_NAME": "hi", "APP_PORT": "9000"},
		},
		{
			name: "comments and blank lines",
			src:  "# top\n\nAPP_A=1 # trailing\n  # indented\nAPP_B=2\n",
			want: map[string]string{"APP_A": "1", "APP_B": "2"},
		},
		{
			name: "whitespace around key and equals",
			src:  "  SPACED_KEY   =   spaced value  \n",
			want: map[string]string{"SPACED_KEY": "spaced value"},
		},
		{
			name: "empty values",
			src:  "EMPTY=\nHASH=# not a value\n",
			want: map[string]string{"EMPTY": "", "HASH": ""},
		},
		{
			name: "hash without space is kept",
			src:  "COLOR=red#blue\n",
			want: map[string]string{"COLOR": "red#blue"},
		},
		{
			name: "single quotes literal",
			src:  "SQ='a $NOT_EXPANDED # kept'\n",
			want: map[string]string{"SQ": "a $NOT_EXPANDED # kept"},
		},
		{
			name: "single quotes with trailing comment",
			src:  "SQ='abc' # comment\n",
			want: map[string]string{"SQ": "abc"},
		},
		{
			name: "double quotes escapes and expansion",
			src:  "DQ=\"a\\n\\\"q\\\"\\\\b $INNER\"\nINNER=in\n",
			want: map[string]string{"DQ": "a\n\"q\"\\b ", "INNER": "in"},
		},
		{
			name: "double quotes with trailing comment",
			src:  "DQ=\"abc\" # comment\n",
			want: map[string]string{"DQ": "abc"},
		},
		{
			name: "export prefix",
			src:  "export FOO=bar\nexport\tBAR=baz\n",
			want: map[string]string{"FOO": "bar", "BAR": "baz"},
		},
		{
			name: "exported is a key not a prefix",
			src:  "exported=1\n",
			want: map[string]string{"exported": "1"},
		},
		{
			name: "expansion dollar and braces prefer OS",
			src:  "FILE_VAR=file\nA=${OS_VAR}-$FILE_VAR-$MISSING",
			want: map[string]string{"FILE_VAR": "file", "A": "os-file-"},
			setup: map[string]string{
				"OS_VAR": "os",
			},
		},
		{
			name: "expansion of earlier keys",
			src:  "BASE=hello\nDERIVED=${BASE}-world\n",
			want: map[string]string{"BASE": "hello", "DERIVED": "hello-world"},
		},
		{
			name: "duplicate keys last wins",
			src:  "DUP=first\nDUP=second\n",
			want: map[string]string{"DUP": "second"},
		},
		{
			name: "multiline double quoted",
			src:  "ML=\"line1\nline2\"\nAFTER=ok\n",
			want: map[string]string{"ML": "line1\nline2", "AFTER": "ok"},
		},
		{
			name: "multiline with escapes",
			src:  "ML=\"a\\nb\"\n",
			want: map[string]string{"ML": "a\nb"},
		},
		{
			name: "crlf endings",
			src:  "A=1\r\nB=2\r\n",
			want: map[string]string{"A": "1", "B": "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.setup {
				t.Setenv(k, v)
			}
			got, err := Parse([]byte(tt.src))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOSBeatsFileInExpansion(t *testing.T) {
	t.Setenv("SHARED", "from-os")
	got, err := Parse([]byte("REF=${SHARED}\nSHARED=from-file\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got["REF"] != "from-os" {
		t.Fatalf("got %q, want from-os", got["REF"])
	}
}

func TestParseMalformed(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantLine int
	}{
		{"missing equals", "OK=1\nNOEQUALS\n", 2},
		{"invalid key", "9BAD=x\n", 1},
		{"unterminated double", "A=\"oops\n", 1},
		{"unterminated single", "A='oops\n", 1},
		{"junk after quote", "A=\"x\" junk\n", 1},
		{"junk after single", "A='x' junk\n", 1},
		{"unterminated multiline", "A=\"one\ntwo\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.src))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var perr *ParseError
			if !errors.As(err, &perr) {
				t.Fatalf("expected *ParseError, got %T (%v)", err, err)
			}
			if perr.Line != tt.wantLine {
				t.Fatalf("line = %d, want %d (%v)", perr.Line, tt.wantLine, err)
			}
		})
	}
}

func TestParseDoesNotTouchOSEnv(t *testing.T) {
	t.Setenv("APP_PORT", "7000")
	m, err := Parse([]byte("APP_PORT=8000\n"))
	if err != nil {
		t.Fatal(err)
	}
	if m["APP_PORT"] != "8000" {
		t.Fatalf("parse wrong: %v", m)
	}
	if os.Getenv("APP_PORT") != "7000" {
		t.Fatal("Parse must not touch OS env")
	}
}
